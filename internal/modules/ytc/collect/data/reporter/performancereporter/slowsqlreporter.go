package performancereporter

import (
	"encoding/json"
	"fmt"
	"sort"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/performance"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
)

type genReportFunc func(titlePrefix string, index int, slowSQL datadef.YTCItem) (content reporter.ReportContent, err error)

// validate interface
var _ commons.Reporter = (*SlowSqlReporter)(nil)

type SlowSqlReporter struct{}

func NewSlowSqlReporter() SlowSqlReporter {
	return SlowSqlReporter{}
}

// [Interface Func]
func (r SlowSqlReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, performance.PerformanceChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}
	// report slow sql
	titleContent := r.genTitleRepoterCentent(title, fontSize)
	content.Txt += titleContent.Txt + stringutil.STR_NEWLINE
	content.Markdown += titleContent.Markdown + stringutil.STR_NEWLINE
	content.HTML += titleContent.HTML + stringutil.STR_NEWLINE

	childs, childFunc := r.genChildrenReportFunc(item)
	for i, c := range childs {
		var childContent reporter.ReportContent
		childContent, err = childFunc[c](titlePrefix, i+1, item)
		if err != nil {
			return
		}
		content.Txt += childContent.Txt + stringutil.STR_NEWLINE
		content.HTML += childContent.HTML + stringutil.STR_NEWLINE
		content.Markdown += childContent.Markdown + stringutil.STR_NEWLINE
	}
	return
}

func (r SlowSqlReporter) genTitleRepoterCentent(title string, font reporter.FontSize) (content reporter.ReportContent) {
	content = reporter.GenReportContentByTitle(title, font)
	return
}

func (r SlowSqlReporter) genExistChildren(slowSQL datadef.YTCItem) (res []string) {
	defer sort.Strings(res)
	chrids := []string{
		performance.KEY_SLOW_SQL_PARAMETER,
		performance.KEY_SLOW_SQL_LOGS,
		performance.KEY_SLOW_SQL_CUT_FILE,
	}
	for _, c := range chrids {
		_, ok := slowSQL.Children[c]
		if ok {
			res = append(res, c)
		}
	}
	return
}

func (r SlowSqlReporter) genChildrenReportFunc(slowSQL datadef.YTCItem) (exists []string, indexFunMap map[string]genReportFunc) {
	indexFunMap = make(map[string]genReportFunc)
	exists = r.genExistChildren(slowSQL)
	allFuncs := r.allChildrenReportFunc()
	for _, exist := range exists {
		indexFunMap[exist] = allFuncs[exist]
	}
	return
}

func (r SlowSqlReporter) allChildrenReportFunc() (funMap map[string]genReportFunc) {
	return map[string]genReportFunc{
		performance.KEY_SLOW_SQL_PARAMETER: r.genSlowParamReportCentent,
		performance.KEY_SLOW_SQL_LOGS:      r.genSlowLogsContent,
		performance.KEY_SLOW_SQL_CUT_FILE:  r.genCutSlowFileReportCentent,
	}
}

func (r SlowSqlReporter) genSlowParamReportCentent(titlePrefix string, index int, slowSQL datadef.YTCItem) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.%d %s", titlePrefix, index, performance.PerformanceChildChineseName[performance.KEY_SLOW_SQL_PARAMETER])
	fontSize := reporter.FONT_SIZE_H3
	slowParameter, ok := slowSQL.Children[performance.KEY_SLOW_SQL_PARAMETER]
	if !ok {
		return
	}
	if !stringutil.IsEmpty(slowParameter.Error) {
		ew := commons.ReporterWriter.NewErrorWriter(slowParameter.Error, slowParameter.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}
	parameters, err := r.parseSlowParameter(slowParameter)
	if err != nil {
		return
	}
	writer := r.genSlowSQLParameterContentWriter(parameters)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r SlowSqlReporter) genSlowSQLParameterContentWriter(parameters []*yasdb.VParameter) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"慢SQL参数名称", "参数值"})
	for _, p := range parameters {
		tw.AppendRow(table.Row{p.Name, p.Value})
	}
	return tw
}

func (r SlowSqlReporter) genCutSlowFileReportCentent(titlePrefix string, index int, slowSQL datadef.YTCItem) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.%d %s", titlePrefix, index, performance.PerformanceChildChineseName[performance.KEY_SLOW_SQL_CUT_FILE])
	fontSize := reporter.FONT_SIZE_H3
	curFileItem, ok := slowSQL.Children[performance.KEY_SLOW_SQL_CUT_FILE]
	if !ok {
		return
	}
	if !stringutil.IsEmpty(curFileItem.Error) {
		ew := commons.ReporterWriter.NewErrorWriter(curFileItem.Error, curFileItem.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}
	cutFile, err := r.parseCutSlowFile(curFileItem)
	if err != nil {
		return
	}
	writer := r.genCutSlowFileWriter(cutFile)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r SlowSqlReporter) genSlowLogsContent(titlePrefix string, index int, slowSQL datadef.YTCItem) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.%d %s", titlePrefix, index, performance.PerformanceChildChineseName[performance.KEY_SLOW_SQL_LOGS])
	fontSize := reporter.FONT_SIZE_H3
	slowLogItem, ok := slowSQL.Children[performance.KEY_SLOW_SQL_LOGS]
	if !ok {
		return
	}
	if !stringutil.IsEmpty(slowLogItem.Error) {
		ew := commons.ReporterWriter.NewErrorWriter(slowLogItem.Error, slowLogItem.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}
	slowLogs, err := r.parseSlowLogs(slowLogItem)
	if err != nil {
		return
	}
	writer := r.genSlowLogWriter(slowLogs)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r SlowSqlReporter) genSlowLogWriter(slowLogs []*yasdb.SlowLog) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"DATABASE_NAME", "USER_NAME", "START_TIME", "USER_HOST", "QUERY_TIME", "ROWS_SENT", "SQL_ID", "SQL_TEXT"})
	for _, slowLog := range slowLogs {
		tw.AppendRow(table.Row{slowLog.DBName, slowLog.UserName, slowLog.StartTime, slowLog.UserHost, slowLog.QueryTime, slowLog.RowsSent, slowLog.SQLID, slowLog.SQLText})
	}
	return tw
}

func (r SlowSqlReporter) genCutSlowFileWriter(cutFile string) reporter.Writer {
	return commons.GenPathWriter(cutFile)
}

func (r SlowSqlReporter) parseSlowParameter(slowParameter datadef.YTCItem) (parameters []*yasdb.VParameter, err error) {
	parameters, ok := slowParameter.Details.([]*yasdb.VParameter)
	if !ok {
		tmp, ok := slowParameter.Details.([]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: performance.KEY_SLOW_SQL_PARAMETER,
				Targets: []interface{}{
					[]*yasdb.VParameter{},
					[]interface{}{},
				},
				Current: slowParameter.Details,
			}
			err = yaserr.Wrapf(err, "parse slow parameter")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &parameters); err != nil {
			err = yaserr.Wrapf(err, "unmarshal slow parameter")
			return
		}
	}
	return
}

func (r SlowSqlReporter) parseCutSlowFile(cutFileItem datadef.YTCItem) (curFile string, err error) {
	curFile, err = commons.ParseString(performance.KEY_SLOW_SQL_CUT_FILE, cutFileItem.Details, "parse slow log cut file")
	return
}

func (r SlowSqlReporter) parseSlowLogs(slowLogsItem datadef.YTCItem) (logs []*yasdb.SlowLog, err error) {
	logs, ok := slowLogsItem.Details.([]*yasdb.SlowLog)
	if !ok {
		tmp, ok := slowLogsItem.Details.([]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: performance.KEY_SLOW_SQL_LOGS,
				Targets: []interface{}{
					[]interface{}{},
					[]*yasdb.SlowLog{},
				},
				Current: slowLogsItem.Details,
			}
			err = yaserr.Wrapf(err, "parse slow logs")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &logs); err != nil {
			err = yaserr.Wrapf(err, "unmarshal slow logs")
			return
		}
	}
	return
}
