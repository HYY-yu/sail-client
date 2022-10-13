package sail

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
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
