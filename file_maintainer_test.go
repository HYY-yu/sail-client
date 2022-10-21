package sail

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestFileMaintainer_saveConfigFile(t *testing.T) {
	tests := []struct {
		name        string
		configs     []string
		response    *clientv3.GetResponse
		mergeConfig bool
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
			mergeConfig: false,
		},
		{
			name:    "TEST2",
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
			mergeConfig: true,
		},
	}
	sail := New(&MetaConfig{
		ETCDEndpoints: "127.0.0.1:2379",
		LogLevel:      "DEBUG",
		ProjectKey:    "test_project_key",
		Namespace:     "test",
		NamespaceKey:  "NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD",
	})

	tempTest, err := os.MkdirTemp("./test_data", "")
	require.NoError(t, err)
	md, err := os.Create(filepath.Join(tempTest, "must_delete.toml"))
	require.NoError(t, err)
	defer func() {
		err = os.RemoveAll(tempTest)
		require.NoError(t, err)
	}()
	sail.metaConfig.ConfigFilePath = tempTest

	_, _ = md.WriteString("CA")
	_ = md.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail.configs = tt.configs
			sail.metaConfig.MergeConfig = tt.mergeConfig
			sail.etcdClient = &clientv3.Client{
				KV: &mockKV{KV: clientv3.NewKVFromKVClient(nil, nil), response: tt.response},
			}
			err := sail.pullETCDConfig()
			require.NoError(t, err)

			dirFiles, err := fileutil.ReadDir(tempTest)
			require.NoError(t, err)

			if tt.mergeConfig {
				assert.Contains(t, dirFiles, MergeConfigName)
				return
			}

			for _, e := range dirFiles {
				t.Logf("Find config file: %s", e)
				assert.Contains(t, sail.configs, e)

				assert.NotEqual(t, "must_delete.toml", e)
			}
		})
	}
}
