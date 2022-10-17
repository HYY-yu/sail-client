package sail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSail_MergeVipersWithName(t *testing.T) {
	tests := []struct {
		name  string
		key   []string
		value []interface{}
	}{
		{
			name: "TEST1",
			key: []string{
				"mysql.yaml.host",
				"redis.properties.host",
				"mysql.yaml.username",
				"mysql.yaml.db_config.maxConnections",
				"temp.custom.temp.custom",
				"test.toml.sail.etcd_username",
			},
			value: []interface{}{
				"127.0.0.1:3306",
				"0.0.0.0",
				"root",
				50,
				"ca",
				"root",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail := initSail(t)

			vipe, err := sail.MergeVipersWithName()
			require.NoError(t, err)

			for i, e := range tt.key {
				v := vipe.Get(e)
				assert.Equal(t, tt.value[i], v)
			}
		})
	}
}

func TestSail_MergeVipers(t *testing.T) {
	tests := []struct {
		name  string
		key   []string
		value []interface{}
	}{
		{
			name: "TEST1",
			key: []string{
				"host",
				"username",
				"db_config.maxConnections",
				"temp.custom",
				"sail.etcd_username",
			},
			value: []interface{}{
				"0.0.0.0",
				"root",
				50,
				"ca",
				"root",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail := initSail(t)

			vipe, err := sail.MergeVipers()
			require.NoError(t, err)

			for i, e := range tt.key {
				v := vipe.Get(e)
				assert.Equal(t, tt.value[i], v)
			}
		})
	}
}

func TestSail_GetWithName(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		key        string
		value      interface{}
	}{
		{
			name:       "TEST1",
			configFile: "mysql.yaml",
			key:        "host",
			value:      "127.0.0.1:3306",
		},
		{
			name:       "TEST2",
			configFile: "mysql.yaml",
			key:        "z_index",
			value:      []interface{}{10, 10, 10},
		},
		{
			name:       "TEST3",
			configFile: "temp.custom",
			key:        "temp.custom",
			value:      "ca",
		},
		{
			name:       "TEST4",
			configFile: "test.toml",
			key:        "sail.etcd_username",
			value:      "root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail := initSail(t)

			v := sail.GetWithName(tt.key, tt.configFile)
			assert.Equal(t, tt.value, v)
		})
	}
}

func TestSail_Get(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "TEST1",
			key:     "host",
			value:   "",
			wantErr: true,
		},
		{
			name:    "TEST2",
			key:     "db_config.maxConnections",
			value:   50,
			wantErr: false,
		},
		{
			name:    "TEST3",
			key:     "port",
			value:   "6379",
			wantErr: false,
		},
		{
			name:    "TEST4",
			key:     "temp.custom",
			value:   "ca",
			wantErr: false,
		},
		{
			name:    "TEST5",
			key:     "sail.log_level",
			value:   "WARN",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail := initSail(t)

			v, err := sail.Get(tt.key)
			if err != nil {
				if tt.wantErr {
					ee, _ := err.(*GetError)
					assert.Equal(t, ErrDuplicateKey, ee.Err)
					return
				}
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.value, v)
		})
	}
}

func initSail(t *testing.T) *Sail {
	sail := New(&MetaConfig{
		ConfigFilePath: "./test_data",
		ETCDEndpoints:  "127.0.0.1:2379",
		LogLevel:       "DEBUG",
		ProjectKey:     "test_project_key",
		Namespace:      "test",
		NamespaceKey:   "NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD",
		Configs:        "mysql.yaml,redis.properties,temp.custom,test.toml",
	})
	err := sail.readLocalFileConfig()
	require.NoError(t, err)

	return sail
}
