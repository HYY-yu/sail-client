package sail

import (
	"os"

	"github.com/spf13/pflag"
)

func NewWithFlag(opts ...Option) *Sail {
	meta, err := getMetaConfigFormFlag()
	if err != nil {
		return &Sail{
			err: err,
		}
	}
	return New(meta, opts...)
}

func getMetaConfigFormFlag() (*MetaConfig, error) {
	meta := MetaConfig{}

	pflag.StringVar(&meta.ETCDEndpoints, "sail-etcd-endpoints", "", "")
	pflag.StringVar(&meta.ETCDUsername, "sail-etcd-username", "", "")
	pflag.StringVar(&meta.ETCDPassword, "sail-etcd-password", "", "")
	pflag.StringVar(&meta.ProjectKey, "sail-project-key", "", "")
	pflag.StringVar(&meta.Namespace, "sail-namespace", "", "")
	pflag.StringVar(&meta.NamespaceKey, "sail-namespace-key", "", "")
	pflag.StringVar(&meta.Configs, "sail-configs", "", "")
	pflag.StringVar(&meta.ConfigFilePath, "sail-config-file-path", "", "")
	pflag.StringVar(&meta.LogLevel, "sail-log-level", "", "")
	pflag.BoolVar(&meta.MergeConfig, "sail-merge-config", false, "")

	err := pflag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	return &meta, nil
}
