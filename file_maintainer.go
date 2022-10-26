package sail

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
)

// FileMaintainer
// 2. 如果设置了ConfigPath，检查Path里有没有配置文件（如果设置了mergeConfig还要merge），有则更新、新增、删除等。(研究是否可以利用Viper)
// 1. pullETCDConfig后，把viper内的配置全部写成文件。
// 2. 如果设置mergeConfig，则mergeViper后，再写成文件。
// 3. 有watch事件，把对应viper的配置重新写成文件。
// 4. 有mergeConfig，重新mergeViper，覆盖写。
// 文件默认不加密
type FileMaintainer struct {
	sail *Sail

	ctx context.Context
}

func NewFileMaintainer(sail *Sail) *FileMaintainer {
	return &FileMaintainer{
		sail: sail,
		ctx:  sail.ctx,
	}
}

func (f *FileMaintainer) saveConfigFile() error {
	if len(f.sail.metaConfig.ConfigFilePath) == 0 {
		return nil
	}

	if f.sail.metaConfig.MergeConfig {
		mergeViper, err := f.sail.MergeVipersWithName()
		if err != nil {
			return err
		}
		err = mergeViper.WriteConfigAs(filepath.Join(f.sail.metaConfig.ConfigFilePath, MergeConfigName))
		if err != nil {
			return err
		}
		return nil
	}

	// 删掉原先的配置
	dirFiles, err := fileutil.ReadDir(f.sail.metaConfig.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("read config file path err: %w ", err)
	}
	deleteConfigFileMap := make(map[string]int)
	for i, e := range dirFiles {
		deleteConfigFileMap[e] = i
	}

	f.sail.lock.RLock()
	defer f.sail.lock.RUnlock()
	for k, v := range f.sail.vipers {
		err := v.WriteConfigAs(filepath.Join(f.sail.metaConfig.ConfigFilePath, k))
		if err != nil {
			return err
		}
		_, ok := deleteConfigFileMap[k]
		if ok {
			delete(deleteConfigFileMap, k)
		}
	}

	for k := range deleteConfigFileMap {
		err := os.Remove(filepath.Join(f.sail.metaConfig.ConfigFilePath, k))
		if err != nil {
			// 没删掉，也不影响正常运行
			f.sail.l.Warn("can't delete file. ", "config_file", k)
		}
	}
	return nil
}

func (f *FileMaintainer) asyncWriteConfigFile(configFileKey string) {
	go func() {
		f.sail.lock.RLock()
		defer f.sail.lock.RUnlock()

		v, ok := f.sail.vipers[configFileKey]
		if !ok {
			return
		}

		err := v.WriteConfigAs(filepath.Join(f.sail.metaConfig.ConfigFilePath, configFileKey))
		if err != nil {
			f.sail.l.Error("refresh config file fail. ", "config_file", configFileKey)
		}
	}()
}
