package fileutil

import (
	"io/fs"
	"os"
)

const (
	DEFAULT_FILE_MODE fs.FileMode = 0644
)

func WriteFile(fname string, data []byte) error {
	return os.WriteFile(fname, data, DEFAULT_FILE_MODE)
}
