package sail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSail_readLocalFileConfig(t *testing.T) {
	sail := New(&MetaConfig{
		ConfigFilePath: "./test_data",
		ETCDEndpoints:  "127.0.0.1:2379",
		LogLevel:       "DEBUG",
		ProjectKey:     "test_project_key",
		Namespace:      "test",
		NamespaceKey:   "NTUZNTNQNUKYEL4GP5SGVDV9LEYZAWBD",
	})
	tests := []struct {
		name        string
		mergeConfig bool
		configs     []string
		keys        []string
		content     []string
	}{
		{
			name:        "TEST1",
			mergeConfig: false,
			configs:     []string{"temp.custom", "test.toml"},
			keys:        []string{"temp.custom", "sail.project_key"},
			content:     []string{"ca", "8a1b491062690963bd978fb8a6958371"},
		}, {
			name:        "TESTEncrypt",
			mergeConfig: false,
			configs:     []string{"redis.properties"},
			keys:        []string{"host"},
			content:     []string{"0.0.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sail.configs = tt.configs
			err := sail.readLocalFileConfig()
			require.NoError(t, err)

			for i, e := range tt.configs {
				viperS, ok := sail.vipers[e]
				assert.Equal(t, true, ok)

				data := viperS.GetString(tt.keys[i])
				assert.Equal(t, tt.content[i], data)
			}
		})
	}
}
