package performancereporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/performance"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*AWRReporter)(nil)

type AWRReporter struct{}

func NewAWRReporter() AWRReporter {
	return AWRReporter{}
}

// [Interface Func]
func (r AWRReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, performance.PerformanceChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report awr
	awrPath, err := commons.ParseString(item.Name, item.Details, "parse awr path")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(awrPath)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r AWRReporter) genReportContentWriter(awr string) reporter.Writer {
	return commons.GenPathWriter(awr)
}
