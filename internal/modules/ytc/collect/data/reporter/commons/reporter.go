package commons

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

var ReporterWriter = reporter.NewReporterWriter()

type Reporter interface {
	Report(item datadef.YTCItem, titlePrefix string) (reporter.ReportContent, error)
}
