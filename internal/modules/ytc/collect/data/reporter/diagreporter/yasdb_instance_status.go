package diagreporter

import (
	"encoding/json"
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/internal/modules/ytc/collect/yasdb"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
)

// validate interface
var _ commons.Reporter = (*YashanDBInstanceStatusReporter)(nil)

type YashanDBInstanceStatusReporter struct{}

func NewYashanDBInstanceStatusReporter() YashanDBInstanceStatusReporter {
	return YashanDBInstanceStatusReporter{}
}

// [Interface Func]
func (r YashanDBInstanceStatusReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report yasdb instance status
	instance, err := r.parseYashanDBVInstance(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse yasdb v$instance")
		return
	}
	writer := r.genReportContentWriter(instance)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r YashanDBInstanceStatusReporter) parseYashanDBVInstance(item datadef.YTCItem) (instance *yasdb.VInstance, err error) {
	instance, ok := item.Details.(*yasdb.VInstance)
	if !ok {
		tmp, ok := item.Details.(map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: item.Name,
				Targets: []interface{}{
					&yasdb.VInstance{},
					map[string]interface{}{},
				},
				Current: item.Details,
			}
			err = yaserr.Wrapf(err, "parse instance interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &instance); err != nil {
			err = yaserr.Wrapf(err, "unmarshal instance info")
			return
		}
	}
	return
}

func (r YashanDBInstanceStatusReporter) genReportContentWriter(instance *yasdb.VInstance) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"状态"})

	tw.AppendRow(table.Row{instance.Status})
	return tw
}
