package collect

import (
	"ytc/internal/modules/ytcctl"
	"ytc/log"
)

func Demo(outputDir, reportType string, types map[string]struct{}) (string, error) {
	path, err := ytcctl.Demo(outputDir, reportType, types)
	if err != nil {
		log.Handler.Errorf("ytcctl call demo failed: %s", err)
	}
	return path, err
}
