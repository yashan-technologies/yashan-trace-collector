package resultgenner

import (
	"encoding/json"

	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/utils/fileutil"
)

type Genner interface {
	GenData(data interface{}, path string) error
	GenReport() (reporter.ReportContent, error)
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
