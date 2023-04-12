package sail

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestNewWithEnv(t *testing.T) {
	// SET ENV
	//export SAIL_ETCD_ENDPOINTS=127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379
	//export SAIL_ETCD_USERNAME=root
	//export SAIL_ETCD_Password=root
	//export SAIL_PROJECT_KEY=8a1b491062690963bd978fb8a6958371
	//export SAIL_NAMESPACE=dev
	//export SAIL_CONFIGS=cfg.properties,mysql.toml,redis.yaml
	//export SAIL_CONFIG_FILE_PATH=.
	//export SAIL_LOG_LEVEL=WARN
	_ = os.Setenv("SAIL_ETCD_ENDPOINTS", "127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379")
	_ = os.Setenv("SAIL_ETCD_USERNAME", "root")
	_ = os.Setenv("SAIL_ETCD_Password", "root")
	_ = os.Setenv("SAIL_PROJECT_KEY", "8a1b491062690963bd978fb8a6958371")
	_ = os.Setenv("SAIL_NAMESPACE", "test")
	_ = os.Setenv("SAIL_NAMESPACE_KEY", "NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD")
	_ = os.Setenv("SAIL_CONFIGS", "cfg.properties,mysql.toml,redis.yaml ")
	_ = os.Setenv("SAIL_CONFIG_FILE_PATH", ".")
	_ = os.Setenv("SAIL_LOG_LEVEL", "WARN")

	t.Run("TestNewEnv", func(t *testing.T) {
		meta := getMetaFormEnv()
		assert.Equal(t, "127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379", meta.ETCDEndpoints)

		got := NewWithEnv()
		assert.NoError(t, got.err)
		assert.NotNil(t, got)

		assert.Equal(t, meta, got.metaConfig)
	})
}

func TestNewWithFlag(t *testing.T) {
	flag := `
--sail-etcd-endpoints=127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379 \
--sail-etcd-username=root --sail-etcd-password=root \
--sail-project-key=8a1b491062690963bd978fb8a6958371 --sail-namespace=dev \
--sail-namespace-key=NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD \
--sail-configs=cfg.properties,mysql.toml,redis.yaml \ 
--sail-config-file-path=. --sail-log-level=WARN`

	flag = strings.ReplaceAll(flag, "\\", "")
	flags := strings.Fields(flag)
	os.Args = append([]string{"test"}, flags...)

	t.Run("TestNewWithFlag", func(t *testing.T) {
		flagReader := pflag.NewFlagSet("flagTest", pflag.ContinueOnError)
		pflag.CommandLine = flagReader

		got := NewWithFlag()
		assert.NoError(t, got.err)
		assert.NotNil(t, got)

		assert.Equal(t, "127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379", got.metaConfig.ETCDEndpoints)
		assert.Equal(t, "root", got.metaConfig.ETCDUsername)
		assert.Equal(t, "root", got.metaConfig.ETCDPassword)
	})
}

func TestNewWithFlagSmall(t *testing.T) {
	flag := `
--sail-etcd-endpoints=127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379 \
--sail-project-key=8a1b491062690963bd978fb8a6958371 --sail-namespace=dev \
`
	flag = strings.ReplaceAll(flag, "\\", "")
	flags := strings.Fields(flag)
	os.Args = append([]string{"test"}, flags...)

	t.Run("TestNewWithFlagSmall", func(t *testing.T) {
		flagReader := pflag.NewFlagSet("flagTest", pflag.ContinueOnError)
		pflag.CommandLine = flagReader

		got := NewWithFlag()
		assert.NoError(t, got.err)
		assert.NotNil(t, got)

		assert.Equal(t, "127.0.0.1:2379,127.0.0.1:12379,127.0.0.1:22379", got.metaConfig.ETCDEndpoints)
	})
}

func TestNewWithToml(t *testing.T) {
	t.Run("TestNewWithToml", func(t *testing.T) {
		got := NewWithToml("./test_data/test.toml")
		assert.NoError(t, got.err)
		assert.NotNil(t, got)

		meta, err := getMetaFormToml("./test_data/test.toml")
		assert.NoError(t, err)

		assert.Equal(t, meta, got.metaConfig)
	})

	t.Run("TestNewWithTomlSmall", func(t *testing.T) {
		got := NewWithToml("./test_data/test_small.toml")
		assert.NoError(t, got.err)
		assert.NotNil(t, got)
		t.Log(got.metaConfig)

		meta, err := getMetaFormToml("./test_data/test_small.toml")
		assert.NoError(t, err)

		assert.Equal(t, meta, got.metaConfig)
	})
}

func Test_intersectionSortStringArr(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				a: []string{"mysql.toml", "redis.yaml"},
				b: []string{"cfg.json", "mysql.toml", "redis.yaml"},
			},
			want: []string{"mysql.toml", "redis.yaml"},
		},
		{
			name: "test2",
			args: args{
				a: []string{"mysql.toml", "redis.yaml"},
				b: []string{},
			},
			want: []string{},
		},
		{
			name: "test3",
			args: args{
				a: []string{"mysql.toml", "redis.yaml"},
				b: []string{"ca.cert"},
			},
			want: []string{},
		},
		{
			name: "test4",
			args: args{
				a: []string{".bachrc", "mysql.toml", "redis.yaml"},
				b: []string{"mysql.toml", "redis.yaml"},
			},
			want: []string{"mysql.toml", "redis.yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intersectionSortStringArr(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("intersectionSortStringArr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSail_getETCDKeyPrefix(t *testing.T) {
	type fields struct {
		metaConfig *MetaConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test1",
			fields: fields{
				metaConfig: &MetaConfig{
					ProjectKey: "test_project_key",
					Namespace:  "test",
				},
			},
			want: "/conf/test_project_key/test/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sail{
				metaConfig: tt.fields.metaConfig,
			}
			if got := s.getETCDKeyPrefix(); got != tt.want {
				t.Errorf("getETCDKeyPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getConfigFileKeyFrom(t *testing.T) {
	type args struct {
		etcdKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test",
			args: args{
				etcdKey: "/conf/project_key/test/mysql.toml",
			},
			want: "mysql.toml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getConfigFileKeyFrom(tt.args.etcdKey); got != tt.want {
				t.Errorf("getConfigFileKeyFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSail_pullETCDConfig(t *testing.T) {
	tests := []struct {
		name     string
		configs  []string
		response *clientv3.GetResponse
	}{
		{
			name:    "TEST1",
			configs: []string{"mysql.toml", "redis.properties"},
			response: &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{
						Key:   []byte("/conf/test_project_key/test/mysql.toml"),
						Value: []byte("database=\"127.0.0.1:3306\""),
					},
					// redis.properties
					// host=0.0.0.0
					// port=6379
					{
						Key:   []byte("/conf/test_project_key/test/redis.properties"),
						Value: []byte("I9IfkJSBekxeYbQJSX6zQsvZJwlfj3VyZ6RrtRF4LFI="),
					},
				},
			},
		},
		{
			name:    "TEST2",
			configs: []string{"mysql.toml"},
			response: &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{
						Key:   []byte("/conf/test_project_key/test/mysql.toml"),
						Value: []byte("database=\"127.0.0.1:3306\""),
					},
					{
						Key:   []byte("/conf/test_project_key/test/redis.properties"),
						Value: []byte("I9IfkJSBekxeYbQJSX6zQsvZJwlfj3VyZ6RrtRF4LFI="),
					},
				},
			},
		}, {
			name:    "TEST3",
			configs: []string{"cfg.custom"},
			response: &clientv3.GetResponse{
				Kvs: []*mvccpb.KeyValue{
					{
						Key:   []byte("/conf/test_project_key/test/cfg.custom"),
						Value: []byte("CA"),
					},
				},
			},
		},
	}

	sail := New(&MetaConfig{
		ETCDEndpoints: "127.0.0.1:2379",
		LogLevel:      "DEBUG",
		ProjectKey:    "test_project_key",
		Namespace:     "test",
		NamespaceKey:  "NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail.configs = tt.configs
			sail.etcdClient = &clientv3.Client{
				KV: &mockKV{KV: clientv3.NewKVFromKVClient(nil, nil), response: tt.response},
			}

			if tt.name == "TEST1" {
				sail.vipers = make(map[string]*viper.Viper)
				err := sail.pullETCDConfig()
				assert.NoError(t, err)

				viperC, ok := sail.vipers[tt.configs[0]]
				assert.Equal(t, true, ok)

				database := viperC.GetString("database")
				assert.Equal(t, "127.0.0.1:3306", database)

				viperR, ok := sail.vipers[tt.configs[1]]
				assert.Equal(t, true, ok)

				host := viperR.GetString("host")
				assert.Equal(t, "0.0.0.0", host)

			} else if tt.name == "TEST2" {
				sail.vipers = make(map[string]*viper.Viper)
				err := sail.pullETCDConfig()
				assert.NoError(t, err)

				viperC, ok := sail.vipers[tt.configs[0]]
				assert.Equal(t, true, ok)

				database := viperC.GetString("database")
				assert.Equal(t, "127.0.0.1:3306", database)

				_, ok = sail.vipers["redis.properties"]
				assert.Equal(t, false, ok)
			} else if tt.name == "TEST3" {
				sail.vipers = make(map[string]*viper.Viper)
				err := sail.pullETCDConfig()
				assert.NoError(t, err)

				viperC, ok := sail.vipers[tt.configs[0]]
				assert.Equal(t, true, ok)

				ca := viperC.GetString(tt.configs[0])
				assert.Equal(t, "CA", ca)
			}
		})
	}
}

type mockKV struct {
	clientv3.KV
	response *clientv3.GetResponse
}

func (kv *mockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return kv.response, nil
}

func TestSail_checkPublish(t *testing.T) {
	type args struct {
		etcdValue []byte
	}
	tests := []struct {
		name          string
		args          args
		wantIsPublish bool
		wantReversion int
	}{
		{
			name:          "test1",
			args:          struct{ etcdValue []byte }{etcdValue: []byte("PUBLISH&THIS_IS_TOKEN&1&22&SecretData==")},
			wantIsPublish: true,
			wantReversion: 22,
		},
		{
			name:          "test2",
			args:          struct{ etcdValue []byte }{etcdValue: []byte("")},
			wantIsPublish: false,
			wantReversion: 0,
		},
	}
	sail := New(&MetaConfig{
		ETCDEndpoints: "127.0.0.1:2379",
		LogLevel:      "DEBUG",
		ProjectKey:    "test_project_key",
		Namespace:     "test",
		NamespaceKey:  "NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsPublish, gotReversion := sail.checkPublish(tt.args.etcdValue)
			assert.Equalf(t, tt.wantIsPublish, gotIsPublish, "checkPublish(%v)", tt.args.etcdValue)
			assert.Equalf(t, tt.wantReversion, gotReversion, "checkPublish(%v)", tt.args.etcdValue)
		})
	}
}
