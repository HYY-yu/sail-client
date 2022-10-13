package sail

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/HYY-yu/sail-client/logger"
)

const MergeConfigName = "config.toml"

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
			dirFiles, err := fileutil.ReadDir(s.metaConfig.ConfigFilePath)
			if err != nil {
				return fmt.Errorf("read config file path err: %w ", err)
			}
			// 过滤，只访问 s.configs 内有的
			configFiles := intersectionSortStringArr(dirFiles, s.configs)
			if s.metaConfig.MergeConfig {
				configFiles = []string{MergeConfigName}
			}

			s.lock.Lock()
			for _, e := range configFiles {
				viperFile := viper.New()
				fileSp := strings.Split(e, ".")
				if len(fileSp) != 2 {
					continue
				}
				name := fileSp[0]
				ext := fileSp[1]
				if ext == "custom" {
					// viper 不支持，以文件名：文件内容形式塞到viper
					fileContent, err := os.ReadFile(filepath.Join(s.metaConfig.ConfigFilePath, e))
					if err != nil {
						return fmt.Errorf("can't read local file: %s with unknow err: %w ", e, err)
					}
					viperFile.Set(e, string(fileContent))
				} else {
					viperFile.AddConfigPath(s.metaConfig.ConfigFilePath)
					viperFile.SetConfigName(name)
					err = viperFile.ReadInConfig()
					if err != nil {
						return fmt.Errorf("can't read local file: %s with unknow err: %w ", e, err)
					}
				}
				s.vipers[e] = viperFile
			}
			s.lock.Unlock()

			// 重连 ETCD
			go s.reconnectEtcd()
			return nil
		}
		return fmt.Errorf("can't connect etcd with unknow err: %w ", err)
	} else {
		s.etcdClient = etcdClient
	}
	if s.etcdClient == nil {
		return nil
	}



	// 1. 访问ETCD，取对应配置文件
	// 3. 把读取的配置灌入 viper 中，包装viper，对外提供配置读取服务。
	// 按 每个config一个Viper的方式灌入，读取时，先设置好读哪个config文件。
	// 如果不设置，则循环查找。
	// Custom 类型的配置，请使用一个特殊的key：configname.configtype 获取

	// 其它：
	// fileMaintainer
	// 2. 如果设置了ConfigPath，检查Path里有没有配置文件（如果设置了mergeConfig还要merge），有则更新、新增、删除等。(研究是否可以利用Viper)
	//      这个要起个线程，后台弄比较好
	// ETCDWatcher
	// 3. 与ETCD建立长链接，WatchKey变化（这个一定要自己做）
	//      这个肯定要线程

	return nil
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
			return
		}
	}
}

func (s *Sail) etcdConnect() (*clientv3.Client, error) {
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
