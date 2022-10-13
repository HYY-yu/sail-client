package sail

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml/v2"
)

func NewWithToml(tomlFilePath string, opts ...Option) *Sail {
	meta, err := getMetaFormToml(tomlFilePath)
	if err != nil {
		return &Sail{
			err: err,
		}
	}
	return New(meta, opts...)
}

func getMetaFormToml(tomlFilePath string) (*MetaConfig, error) {
	tomlFile, err := ioutil.ReadFile(tomlFilePath)
	if err != nil {
		return nil, fmt.Errorf("read toml file err: %w ", err)
	}
	type T struct {
		M MetaConfig `toml:"sail"`
	}
	t := T{}

	err = toml.Unmarshal(tomlFile, &t)
	if err != nil {
		return nil, fmt.Errorf("unmarshal toml file err: %w ", err)
	}
	return &t.M, nil
}
