package sail

import (
	"os"

	"github.com/spf13/afero"
)

// Check if file or dir Exists
func exists(fs afero.Fs, path string) (bool, error) {
	_, err := fs.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
