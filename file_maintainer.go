package sail

import (
	"context"
	"path/filepath"
)

type FileMaintainer struct {
	sail *Sail

	ctx context.Context
}

func NewFileMaintainer(sail *Sail) *FileMaintainer {
	ctx, _ := context.WithCancel(sail.ctx)
	return &FileMaintainer{
		sail: sail,
		ctx:  ctx,
	}
}

func (f *FileMaintainer) saveConfigFile() {
	for k, v := range f.sail.vipers {
		v.WriteConfigAs(filepath.Join(f.sail.metaConfig.ConfigFilePath + k))
	}
}
