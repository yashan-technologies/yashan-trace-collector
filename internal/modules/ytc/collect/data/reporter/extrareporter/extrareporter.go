// package extrareporter is used to generate the extra file reports
package extrareporter

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/extra"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
)

// validate interface
var _ commons.Reporter = (*ExtraFileReporter)(nil)

type ExtraFileReporter struct{}

func NewExtraFileReporter() ExtraFileReporter {
	return ExtraFileReporter{}
}

// [Interface Func]
func (r ExtraFileReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := extra.ExtraChineseName[item.Name]
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report extra file
	extraFile, err := commons.ParseString(item.Name, item.Details, "parse extra file")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(extraFile)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r ExtraFileReporter) genReportContentWriter(extraFile string) reporter.Writer {
	return commons.GenPathWriter(extraFile)
}
