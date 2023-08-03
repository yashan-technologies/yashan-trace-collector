package ytcctl

import "ytc/internal/modules/ytc"

func Demo(outputDir, reportType string, types map[string]struct{}) (string, error) {
	return ytc.Demo(outputDir, reportType, types)
}
