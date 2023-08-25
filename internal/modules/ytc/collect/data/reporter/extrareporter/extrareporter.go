// package extrareporter is used to generate the extra file reports
package extrareporter

import (
	"encoding/json"
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/extra"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
)

// validate interface
var _ commons.Reporter = (*ExtraFileReporter)(nil)

type ExtraFileReporter struct{}

func NewExtraFileReporter() ExtraFileReporter {
	return ExtraFileReporter{}
}

// [Interface Func]
func (r ExtraFileReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, extra.ExtraChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	pathMap, err := r.parsePathMap(item)
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(pathMap)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r ExtraFileReporter) genReportContentWriter(pathMap map[string]string) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"当前路径", "源文件路径"})
	for k, v := range pathMap {
		tw.AppendRow(table.Row{k, v})
	}
	return tw
}

func (r ExtraFileReporter) parsePathMap(extraItem datadef.YTCItem) (pathMap map[string]string, err error) {
	pathMap, ok := extraItem.Details.(map[string]string)
	if !ok {
		tmp, ok := extraItem.Details.(map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: extra.KEY_EXTRA_FILE,
				Targets: []interface{}{
					map[string]string{},
					map[string]interface{}{},
				},
				Current: extraItem.Details,
			}
			err = yaserr.Wrapf(err, "parse extra file path")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &pathMap); err != nil {
			err = yaserr.Wrapf(err, "unmarshal extra path map")
			return
		}
	}
	return
}
