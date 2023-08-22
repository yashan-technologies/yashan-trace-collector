package diagreporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*YashanDBADRLogReporter)(nil)

type YashanDBADRLogReporter struct{}

func NewYashanDBADRLogReporter() YashanDBADRLogReporter {
	return YashanDBADRLogReporter{}
}

// [Interface Func]
func (r YashanDBADRLogReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report yasdb adr log
	adrLog, err := commons.ParseString(item.Name, item.Details, "parse yasdb adr log")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(adrLog)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r YashanDBADRLogReporter) genReportContentWriter(adrLog string) reporter.Writer {
	return commons.GenPathWriter(adrLog)
}
