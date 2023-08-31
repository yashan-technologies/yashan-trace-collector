package diagreporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*YashanDBCoreDumpReporter)(nil)

type YashanDBCoreDumpReporter struct{}

func NewYashanDBCoreDumpReporter() YashanDBCoreDumpReporter {
	return YashanDBCoreDumpReporter{}
}

// [Interface Func]
func (r YashanDBCoreDumpReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report yasdb coredump
	coreDump, err := commons.ParseString(item.Name, item.Details, "parse yasdb coredump")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(coreDump)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r YashanDBCoreDumpReporter) genReportContentWriter(coreDump string) reporter.Writer {
	return commons.GenPathWriter(coreDump)
}
