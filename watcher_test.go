package sail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Test_etcdWatcher_dealETCDMsg(t *testing.T) {
	tests := []struct {
		name           string
		configs        []string
		response       *clientv3.GetResponse
		replaceConfig  string
		replaceContent string
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
			replaceConfig:  "/conf/test_project_key/test/mysql.toml",
			replaceContent: "database=\"0.0.0.0:3306\"\ndb_log=true",
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
			err := sail.pullETCDConfig()
			require.NoError(t, err)

			e := NewWatcher(sail.ctx, sail, sail.getETCDKeyPrefix(), sail.etcdClient)
			ee := e.(*etcdWatcher)

			db, err := sail.GetString("database")
			require.NoError(t, err)
			assert.Equal(t, "127.0.0.1:3306", db)
			port, err := sail.GetString("port")
			require.NoError(t, err)
			assert.Equal(t, "6379", port)

			ee.dealETCDMsg(tt.replaceConfig, []byte(tt.replaceContent))

			db2, err := sail.GetString("database")
			require.NoError(t, err)
			assert.Equal(t, "0.0.0.0:3306", db2)

			log, err := sail.GetBool("db_log")
			require.NoError(t, err)
			assert.Equal(t, true, log)

			port, err = sail.GetString("port")
			require.NoError(t, err)
			assert.Equal(t, "6379", port)
		})
	}
}
