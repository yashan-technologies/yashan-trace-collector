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
var _ commons.Reporter = (*YashanDBDatabaseStatusReporter)(nil)

type YashanDBDatabaseStatusReporter struct{}

func NewYashanDBDatabaseStatusReporter() YashanDBDatabaseStatusReporter {
	return YashanDBDatabaseStatusReporter{}
}

// [Interface Func]
func (r YashanDBDatabaseStatusReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report yasdb database status
	databse, err := r.parseYashanDBVDatabase(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse yasdb v$instance")
		return
	}
	writer := r.genReportContentWriter(databse)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r YashanDBDatabaseStatusReporter) parseYashanDBVDatabase(item datadef.YTCItem) (database *yasdb.VDatabase, err error) {
	database, ok := item.Details.(*yasdb.VDatabase)
	if !ok {
		tmp, ok := item.Details.(map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: item.Name,
				Targets: []interface{}{
					&yasdb.VDatabase{},
					map[string]interface{}{},
				},
				Current: item.Details,
			}
			err = yaserr.Wrapf(err, "parse database interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &database); err != nil {
			err = yaserr.Wrapf(err, "unmarshal database info")
			return
		}
	}
	return
}

func (r YashanDBDatabaseStatusReporter) genReportContentWriter(database *yasdb.VDatabase) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"状态", "开启模式"})

	tw.AppendRow(table.Row{database.Status, database.OpenMode})
	return tw
}
