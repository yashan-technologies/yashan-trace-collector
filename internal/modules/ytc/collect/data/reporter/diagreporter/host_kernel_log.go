package diagreporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*HostKernelLogReporter)(nil)

type HostKernelLogReporter struct{}

func NewHostKernelLogReporter() HostKernelLogReporter {
	return HostKernelLogReporter{}
}

// [Interface Func]
func (r HostKernelLogReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report host dmesg log
	demsgLog, err := commons.ParseString(item.Name, item.Details, "parse host dmesg log")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(demsgLog)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r HostKernelLogReporter) genReportContentWriter(demsgLog string) reporter.Writer {
	return commons.GenPathWriter(demsgLog)
}
