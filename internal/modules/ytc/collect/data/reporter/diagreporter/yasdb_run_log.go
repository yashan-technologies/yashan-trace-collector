package diagreporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*YashanDBRunLogReporter)(nil)

type YashanDBRunLogReporter struct{}

func NewYashanDBRunLogReporter() YashanDBRunLogReporter {
	return YashanDBRunLogReporter{}
}

// [Interface Func]
func (r YashanDBRunLogReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report yasdb run log
	runLog, err := commons.ParseString(item.Name, item.Details, "parse yasdb run log")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(runLog)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r YashanDBRunLogReporter) genReportContentWriter(runLog string) reporter.Writer {
	return commons.GenPathWriter(runLog)
}
