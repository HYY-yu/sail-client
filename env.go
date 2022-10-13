package sail

import (
	"os"
	"strconv"
)

func NewWithEnv(opts ...Option) *Sail {
	return New(getMetaFormEnv(), opts...)
}

func getMetaFormEnv() *MetaConfig {
	meta := MetaConfig{
		ETCDEndpoints:  os.Getenv("SAIL_ETCD_ENDPOINTS"),
		ETCDUsername:   os.Getenv("SAIL_ETCD_USERNAME"),
		ETCDPassword:   os.Getenv("SAIL_ETCD_PASSWORD"),
		ProjectKey:     os.Getenv("SAIL_PROJECT_KEY"),
		Namespace:      os.Getenv("SAIL_NAMESPACE"),
		NamespaceKey:   os.Getenv("SAIL_NAMESPACE_KEY"),
		Configs:        os.Getenv("SAIL_CONFIGS"),
		ConfigFilePath: os.Getenv("SAIL_CONFIG_FILE_PATH"),
		LogLevel:       os.Getenv("SAIL_LOG_LEVEL"),
	}
	meta.MergeConfig, _ = strconv.ParseBool(os.Getenv("SAIL_MERGE_CONFIG"))
	return &meta
}
