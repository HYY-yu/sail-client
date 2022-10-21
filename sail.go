package sail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/HYY-yu/seckill.pkg/pkg/encrypt"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/HYY-yu/sail-client/logger"
)

const MergeConfigName = "config.toml"

type OnConfigChange func(configFileKey string, s *Sail)

type MetaConfig struct {
	ETCDEndpoints string `toml:"etcd_endpoints"` // 逗号分隔的ETCD地址，0.0.0.0:2379,0.0.0.0:12379,0.0.0.0:22379
	ETCDUsername  string `toml:"etcd_username"`
	ETCDPassword  string `toml:"etcd_password"`

	ProjectKey   string `toml:"project_key"`
	Namespace    string `toml:"namespace"`
	NamespaceKey string `toml:"namespace_key"`

	Configs        string `toml:"configs"`          // 逗号分隔的 config_name.config_type，如：mysql.toml,cfg.json,redis.yaml，空代表不下载任何配置
	ConfigFilePath string `toml:"config_file_path"` // 本地配置文件存放路径，空代表不存储本都配置文件
	LogLevel       string `toml:"log_level"`        // 日志级别(DEBUG\INFO\WARN\ERROR)，默认 WARN
	MergeConfig    bool   `toml:"merge_config"`     // 是否合并配置，合并配置则会将同类型的配置合并到一个文件中，需要先设置ConfigFilePath
}

func (m *MetaConfig) SplitETCDEndpoints() []string {
	endpoints := strings.Split(m.ETCDEndpoints, ",")
	return endpoints
}

func (m *MetaConfig) SplitConfigs() []string {
	if len(strings.TrimSpace(m.Configs)) == 0 {
		return []string{}
	}
	configs := strings.Split(m.Configs, ",")
	sort.Strings(configs)
	return configs
}

func (m *MetaConfig) valid() error {
	if len(m.ETCDEndpoints) == 0 {
		return errors.New("please set etcd-endpoints. ")
	}

	endpoints := strings.Split(m.ETCDEndpoints, ",")
	for _, e := range endpoints {
		_, _, err := net.SplitHostPort(e)
		if err != nil {
			return fmt.Errorf("parse endpoints fail: %w ", err)
		}
	}

	if len(m.ProjectKey) == 0 {
		return errors.New("please set project-key. ")
	}

	if len(m.Namespace) == 0 {
		return errors.New("please set namespace. ")
	}

	if !checkLogLevel(m.LogLevel) {
		return errors.New("please set correct log-level. ")
	}

	return nil
}

func checkLogLevel(logLevel string) bool {
	if len(strings.TrimSpace(logLevel)) == 0 {
		return true
	}

	right := false
	for _, e := range []string{"DEBUG", "INFO", "WARN", "ERROR"} {
		if e == logLevel {
			right = true
		}
	}
	return right
}

type Sail struct {
	metaConfig *MetaConfig
	l          logger.Logger

	etcdEndpoints []string
	etcdConfig    *clientv3.Config
	etcdClient    *clientv3.Client

	configs []string

	vipers map[string]*viper.Viper
	lock   *sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	changeFunc OnConfigChange

	fm *FileMaintainer

	err error
}

func New(meta *MetaConfig, opts ...Option) *Sail {
	if err := meta.valid(); err != nil {
		return &Sail{
			err: err,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Sail{
		metaConfig: meta,
		l:          logger.New(),

		etcdEndpoints: meta.SplitETCDEndpoints(),
		configs:       meta.SplitConfigs(),

		vipers: make(map[string]*viper.Viper),
		lock:   &sync.RWMutex{},
		ctx:    ctx,
		cancel: cancel,
	}
	thre := map[string]jww.Threshold{
		"DEBUG": 1,
		"INFO":  2,
		"WARN":  3,
		"ERROR": 4,
	}[s.metaConfig.LogLevel]

	jww.SetStdoutThreshold(thre)

	for _, opt := range opts {
		opt.apply(s)
	}

	s.fm = NewFileMaintainer(s)
	return s
}

type Option interface {
	apply(v *Sail)
}

type optionFunc func(v *Sail)

func (fn optionFunc) apply(v *Sail) {
	fn(v)
}

func WithConfigs(configs []string) Option {
	return optionFunc(func(v *Sail) {
		v.configs = configs
	})
}

func WithConfigPath(configPath string) Option {
	return optionFunc(func(v *Sail) {
		v.metaConfig.ConfigFilePath = configPath
	})
}

func WithOnConfigChange(f OnConfigChange) Option {
	return optionFunc(func(v *Sail) {
		v.changeFunc = f
	})
}

// WithLogger
// TODO 在文档中说明，强烈建议替换为项目自己的Logger，这样可以实时变化 loggerLevel
func WithLogger(logger logger.Logger) Option {
	return optionFunc(func(v *Sail) {
		v.l = logger
	})
}

// WithMergeConfig 将所有的配置都放到一个文件: config.toml
func WithMergeConfig(merge bool) Option {
	return optionFunc(func(v *Sail) {
		v.metaConfig.MergeConfig = merge
	})
}

func WithETCDClientConfig(cfg *clientv3.Config) Option {
	return optionFunc(func(v *Sail) {
		v.etcdConfig = cfg
	})
}

func (s *Sail) Err() error {
	return s.err
}

func (s *Sail) Pull() error {
	if s.Err() != nil {
		return s.Err()
	}

	etcdClient, err := s.etcdConnect()
	if err != nil {
		if err == context.DeadlineExceeded && !fileutil.DirEmpty(s.metaConfig.ConfigFilePath) {
			s.l.Warn("using local file because can't connect etcd, the connection will retry after 30s. ")
			err := s.readLocalFileConfig()
			if err != nil {
				return err
			}

			// 重连 ETCD
			go s.reconnectEtcd()
			return nil
		}
		return fmt.Errorf("can't connect etcd with unknow err: %w ", err)
	}
	s.l.Debug("connect etcd success. ")
	s.etcdClient = etcdClient

	err = s.pullETCDConfig()
	if err != nil {
		return err
	}

	return nil
}

func (s *Sail) pullETCDConfig() error {
	if len(s.configs) == 0 {
		// 不获取任何配置，直接退出
		return nil
	}

	// s.configs 是有序的，etcd里的key是有序的
	// 那么，只需要获取 from-key = s.configs[0]   limit=s.configs 即可获取到所有需要的key。
	keyPrefix := s.getETCDKeyPrefix()
	formKey := s.configs[0]
	limit := len(s.configs)
	s.l.Debug("pull config key", "keys", s.configs)

	getResp, err := s.etcdClient.Get(s.ctx,
		keyPrefix+formKey,
		clientv3.WithFromKey(),
		clientv3.WithLimit(int64(limit)),
	)
	if err != nil {
		return fmt.Errorf("read config from etcd err: %w ", err)
	}
	etcdKeys := make([]string, 0, len(s.configs))
	for _, e := range getResp.Kvs {
		etcdKeys = append(etcdKeys, getConfigFileKeyFrom(string(e.Key)))
	}
	if len(etcdKeys) == 0 {
		return fmt.Errorf("read empty config from etcd! ")
	}

	insETCDKeys := intersectionSortStringArr(etcdKeys, s.configs)
	s.l.Debug("real config key", "keys", insETCDKeys)

	s.lock.Lock()
	for _, e := range getResp.Kvs {
		configFileKey := getConfigFileKeyFrom(string(e.Key))
		for _, ins := range insETCDKeys {
			if ins == configFileKey {
				viperETCD, err := s.newViperWithETCDValue(configFileKey, e.Value)
				if err != nil {
					s.lock.Unlock()
					return nil
				}
				if viperETCD == nil {
					continue
				}

				s.vipers[configFileKey] = viperETCD
			}
		}
	}
	s.lock.Unlock()

	err = s.fm.saveConfigFile()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sail) newViperWithETCDValue(configFileKey string, etcdValue []byte) (*viper.Viper, error) {
	viperETCD := viper.New()
	configType := strings.TrimPrefix(filepath.Ext(configFileKey), ".")
	valueReader := bytes.NewBuffer(etcdValue)

	if c := s.tryDecryptConfigContent(configFileKey, valueReader.String()); len(c) > 0 {
		valueReader = bytes.NewBufferString(c)
	} else {
		s.l.Error("decrypt config fail, skip it. ", "key", configFileKey)
		return nil, nil
	}

	if configType == "custom" {
		viperETCD.Set(configFileKey, valueReader.String())
	} else {
		viperETCD.SetConfigType(configType)
		err := viperETCD.ReadConfig(valueReader)
		if err != nil {
			return nil, fmt.Errorf("viper fail: read config from etcd err: %w ", err)
		}
	}
	return viperETCD, nil
}

func (s *Sail) tryDecryptConfigContent(configKey, content string) string {
	_, err := encrypt.NewBase64Encoding().DecodeString(content)
	if err == nil {
		decryptContent, err := decryptConfigContent(content, s.metaConfig.NamespaceKey)
		if err != nil {
			// 报错、跳过，不中断运行。
			s.l.Error("decrypt config %s err:%w ", configKey, err)
			return ""
		}
		content = decryptContent
	}
	return content
}

func decryptConfigContent(content string, namespaceKey string) (string, error) {
	if namespaceKey == "" {
		return "", nil
	}

	goAES := encrypt.NewGoAES(namespaceKey, encrypt.AES192)
	decryptContent, err := goAES.WithModel(encrypt.ECB).WithEncoding(encrypt.NewBase64Encoding()).Decrypt(content)
	if err != nil {
		return "", err
	}
	return decryptContent, nil
}

// /conf/{project_key}/namespace/config_name.config.type
func (s *Sail) getETCDKeyPrefix() string {
	b := strings.Builder{}

	b.WriteString("/conf")

	b.WriteByte('/')
	b.WriteString(s.metaConfig.ProjectKey)

	b.WriteByte('/')
	b.WriteString(s.metaConfig.Namespace)

	b.WriteByte('/')
	return b.String()
}

func getConfigFileKeyFrom(etcdKey string) string {
	_, result := filepath.Split(etcdKey)
	return result
}

func (s *Sail) reconnectEtcd() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(30 * time.Second):
			etcdClient, err := s.etcdConnect()
			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				s.l.Error("reconnect exit because err: %w ", err)
				return
			}
			s.etcdClient = etcdClient
			s.l.Info("reconnect etcd success! ")
			err = s.pullETCDConfig()
			if err != nil {
				s.l.Error("pull etcd config fail, retry in next 30s. ")
				_ = s.etcdClient.Close()
				s.etcdClient = nil
				continue
			}
			return
		}
	}
}

func (s *Sail) etcdConnect() (*clientv3.Client, error) {
	s.l.Debug("start to connect etcd. ")
	v3cfg := &clientv3.Config{
		Endpoints:            s.etcdEndpoints,
		AutoSyncInterval:     time.Minute,
		DialTimeout:          10 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 20 * time.Second,
		Username:             s.metaConfig.ETCDUsername,
		Password:             s.metaConfig.ETCDPassword,
		PermitWithoutStream:  true,
		DialOptions:          []grpc.DialOption{grpc.WithBlock()},
	}
	if s.etcdConfig != nil {
		s.etcdConfig.Endpoints = v3cfg.Endpoints
		s.etcdConfig.Username = v3cfg.Username
		s.etcdConfig.Password = v3cfg.Password

		v3cfg = s.etcdConfig
	}

	return clientv3.New(*v3cfg)
}

func (s *Sail) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.etcdClient != nil {
		err := s.etcdClient.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func intersectionSortStringArr(a []string, b []string) []string {
	if (len(a) == 0) || (len(b) == 0) {
		return []string{}
	}

	result := make([]string, 0, len(a))
	i, j := 0, 0
	for i != len(a) && j != len(b) {
		if a[i] > b[j] {
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			result = append(result, a[i])
			i++
			j++
		}
	}
	return result
}
