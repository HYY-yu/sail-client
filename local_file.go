package sail

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
)

func (s *Sail) readLocalFileConfig() error {
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
		ext := fileSp[1]
		fileContent, err := os.ReadFile(filepath.Join(s.metaConfig.ConfigFilePath, e))
		if err != nil {
			return fmt.Errorf("can't read local file: %s with unknow err: %w ", e, err)
		}
		fContent := string(fileContent)

		if c := s.tryDecryptConfigContent(e, fContent); len(c) > 0 {
			fContent = c
		} else {
			continue
		}

		if ext == "custom" {
			// viper 不支持的格式，就以文件名：文件内容形式塞到viper
			viperFile.Set(e, fContent)
		} else {
			viperFile.SetConfigType(ext)
			err = viperFile.ReadConfig(bytes.NewBufferString(fContent))
			if err != nil {
				return fmt.Errorf("can't read local file: %s with unknow err: %w ", e, err)
			}
		}
		s.vipers[e] = viperFile
	}
	s.lock.Unlock()
	return nil
}
