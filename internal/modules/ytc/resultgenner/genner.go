package resultgenner

import (
	"encoding/json"
	"ytc/utils/fileutil"
)

type Genner interface {
	GenData(data interface{}, path string) error
	GenReport() (content []byte)
}

type BaseGenner struct{}

func (g BaseGenner) GenData(data interface{}, path string) error {
	bytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return fileutil.WriteFile(path, bytes)
}

func (g BaseGenner) GenReport() []byte {
	return nil
}
