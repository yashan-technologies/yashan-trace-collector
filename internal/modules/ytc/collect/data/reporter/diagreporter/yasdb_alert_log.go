package diagreporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*YashanDBAlertLogReporter)(nil)

type YashanDBAlertLogReporter struct{}

func NewYashanDBAlertLogReporter() YashanDBAlertLogReporter {
	return YashanDBAlertLogReporter{}
}

// [Interface Func]
func (r YashanDBAlertLogReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report yasdb alert log
	alertLog, err := commons.ParseString(item.Name, item.Details, "parse yasdb alert log")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(alertLog)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r YashanDBAlertLogReporter) genReportContentWriter(alertLog string) reporter.Writer {
	return commons.GenPathWriter(alertLog)
}
