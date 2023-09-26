package performance

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	"ytc/defs/timedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/execer"
	"git.yasdb.com/go/yasutil/fs"
)

const (
	VIEW_SLOW_LOG = "SLOW_LOG$"

	// SLOW_LOG in file prefix
	SqlPrefix          = "SQL: "
	UserHostPrefix     = "# USER_HOST: "
	DBNamePrefix       = "# DB_NAME: "
	ExecuteTimePrefix  = "# COST_EXECUTE_TIME: "
	OptimizeTimePrefix = "# COST_OPTIMIZE_TIME: "
	RowsSentPrefix     = "# ROWS_SENT: "
	SqlIDPrefix        = "# SQL_ID: "
	TimePrefix         = "# TIME: "

	// child name
	KEY_SLOW_SQL_PARAMETER     = "slowParameter"
	KEY_SLOW_SQL_LOGS_IN_TABLE = "slowLogsInFile"
	KEY_SLOW_SQL_LOGS_IN_FILE  = "slowCutFile"

	SLOW_LOG_FILE_PATH = "SLOW_LOG_FILE_PATH"
	AWR                = "awr"
	SLOW               = "slowsql"
	YASQL_ERR_PREFIX   = "YAS-"

	// awr sql
	_set_output      = "set serveroutput on"
	_exec_awr_report = "exec sys.dbms_awr.awr_report(%d,%d,%d,%d);"
)

var (
	_packageDir string
)

var (
	PerformanceChineseName = map[string]string{
		datadef.PERF_YASDB_AWR:      "AWR报告",
		datadef.PERF_YASDB_SLOW_SQL: "慢SQL",
	}

	PerformanceChildChineseName = map[string]string{
		KEY_SLOW_SQL_PARAMETER:     "慢SQL参数",
		KEY_SLOW_SQL_LOGS_IN_TABLE: "SLOW_LOG$系统表",
		KEY_SLOW_SQL_LOGS_IN_FILE:  "慢SQL日志文件",
	}

	_slowParameter = []yasdb.ParameterName{
		yasdb.ENABLE_SLOW_LOG,
		yasdb.SLOW_LOG_OUTPUT,
		yasdb.SLOW_LOG_FILE_PATH,
		yasdb.SLOW_LOG_TIME_THRESHOLD,
		yasdb.SLOW_LOG_SQL_MAX_LEN,
	}
)

type PerfCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *datadef.YTCModule
	yasdbValidateErr error
}

type execRes struct {
	ret    int
	stdout string
	stderr string
}

func NewPerfCollecter(collectParam *collecttypedef.CollectParam) *PerfCollecter {
	return &PerfCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &datadef.YTCModule{
			Module: collecttypedef.TYPE_PERF,
		},
	}
}

// [Interface Func]
func (p *PerfCollecter) PreCollect(packageDir string) error {
	p.setPackageDir(packageDir)
	if err := fs.Mkdir(p.getAWRPath()); err != nil {
		return err
	}
	if err := fs.Mkdir(p.getSlowPath()); err != nil {
		return err
	}
	return nil
}

// [Interface Func]
func (p *PerfCollecter) Type() string {
	return collecttypedef.TYPE_PERF
}

// [Interface Func]
func (p *PerfCollecter) CheckAccess(yasdbValidate error) (noAccess []ytccollectcommons.NoAccessRes) {
	p.yasdbValidateErr = yasdbValidate
	noAccess = make([]ytccollectcommons.NoAccessRes, 0)
	funcMap := p.checkFunc()
	for item, fn := range funcMap {
		noAccessRes := fn()
		if noAccessRes != nil {
			log.Module.Debugf("item [%s] check asscess desc: %s tips %s", item, noAccessRes.Description, noAccessRes.Tips)
			noAccess = append(noAccess, *noAccessRes)
		}
	}
	return
}

// [Interface Func]
func (p *PerfCollecter) CollectFunc(items []string) (res map[string]func() error) {
	res = make(map[string]func() error)
	itemFuncMap := p.itemFunc()
	for _, collectItem := range items {
		_, ok := itemFuncMap[collectItem]
		if !ok {
			log.Module.Errorf("get %s collect func err", collectItem)
			continue
		}
		res[collectItem] = itemFuncMap[collectItem]
	}
	return
}

// [Interface Func]
func (p *PerfCollecter) ItemsToCollect(noAccess []ytccollectcommons.NoAccessRes) (res []string) {
	noAccessMap := ytccollectcommons.NotAccessItemToMap(noAccess)
	for item := range PerformanceChineseName {
		if _, ok := noAccessMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

// [Interface Func]
func (p *PerfCollecter) CollectOK() *datadef.YTCModule {
	return p.ModuleCollectRes
}

func (p *PerfCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		datadef.PERF_YASDB_AWR:      p.collectAWR,
		datadef.PERF_YASDB_SLOW_SQL: p.collectSlowSQL,
	}
}

func (b *PerfCollecter) setPackageDir(packageDir string) {
	_packageDir = packageDir

}

func (p *PerfCollecter) checkFunc() map[string]func() *ytccollectcommons.NoAccessRes {
	return map[string]func() *ytccollectcommons.NoAccessRes{
		datadef.PERF_YASDB_AWR:      p.checkAWR,
		datadef.PERF_YASDB_SLOW_SQL: p.checkSlowSql,
	}
}

func (p *PerfCollecter) fillResult(data *datadef.YTCItem) {
	p.ModuleCollectRes.Set(data)
}

func (p *PerfCollecter) getSlowLogPath() (string, error) {
	tx := yasqlutil.GetLocalInstance(p.YasdbUser, p.YasdbPassword, p.YasdbHome, p.YasdbData)
	slowPath, err := yasdb.QueryParameter(tx, SLOW_LOG_FILE_PATH)
	if err != nil {
		return "", err
	}
	slowPath = strings.ReplaceAll(slowPath, "?", p.YasdbData)
	return slowPath, nil
}

func (p *PerfCollecter) collectAWR() error {
	awr := &datadef.YTCItem{Name: datadef.PERF_YASDB_AWR}
	defer p.fillResult(awr)
	log := log.Module.M(datadef.PERF_YASDB_AWR)
	sqlFile, err := p.createAWRSqlFile(log)
	if err != nil {
		awr.Error = err.Error()
		awr.Description = datadef.GenDefaultDesc()
		return err
	}
	defer p.deleteSqlFile(sqlFile)
	htmlFile, err := p.genAWRHtmlReport(log, sqlFile)
	if err != nil {
		awr.Error = err.Error()
		awr.Description = datadef.GenDefaultDesc()
		return err
	}
	relative, err := filepath.Rel(_packageDir, htmlFile)
	if err != nil {
		awr.Error = err.Error()
		awr.Description = datadef.GenDefaultDesc()
		return err
	}
	awr.Details = p.GenPackageRelativePath(relative)
	return nil
}

func (p *PerfCollecter) createAWRSqlFile(log yaslog.YasLog) (string, error) {
	dataInstance, err := p.queryDatabaseInstance(log)
	if err != nil {
		log.Errorf("query database_instance err: %s", err.Error())
		return "", err
	}
	startId, endId, err := p.genStartEndSnapId(log)
	if err != nil {
		log.Errorf("gen snapshot id err: %s", err.Error())
		return "", err
	}
	var buffer bytes.Buffer
	buffer.WriteString(_set_output + stringutil.STR_NEWLINE)
	buffer.WriteString(fmt.Sprintf(_exec_awr_report, dataInstance.DBID, dataInstance.InstanceNumber, startId, endId) + stringutil.STR_NEWLINE)
	awrDir := p.getAWRPath()
	sqlName := fmt.Sprintf("%d-%d-%d-%d.sql", dataInstance.DBID, dataInstance.InstanceNumber, startId, endId)
	sqlFile := path.Join(awrDir, sqlName)
	if err := fileutil.WriteFile(sqlFile, buffer.Bytes()); err != nil {
		return "", err
	}
	log.Infof("gen awr sql: \n%s", buffer.String())
	return sqlFile, nil
}

func (p *PerfCollecter) genStartEndSnapId(log yaslog.YasLog) (int64, int64, error) {
	tx := yasqlutil.GetLocalInstance(p.YasdbUser, p.YasdbPassword, p.YasdbHome, p.YasdbData)

	instance, err := yasdb.QueryInstance(tx)
	if err != nil {
		log.Errorf("query instance failed: %s", err.Error())
		return 0, 0, err
	}
	tx.ResetError()

	instanceStartupTime, err := time.Parse(timedef.TIME_FORMAT, instance.StartupTime)
	if err != nil {
		log.Errorf("parse instance startup time %s failed: %s", instance.StartupTime, err.Error())
		return 0, 0, err
	}
	start := p.StartTime.Format(timedef.TIME_FORMAT)
	if p.StartTime.Before(instanceStartupTime) {
		start = instanceStartupTime.Format(timedef.TIME_FORMAT)
	}

	end := p.EndTime.Format(timedef.TIME_FORMAT)
	snaps, err := yasdb.QueryWrmSnapsot(tx, start, end)
	if err != nil {
		log.Errorf("query snapshot err: %s", err.Error())
		return 0, 0, err
	}
	max, min := int64(math.MinInt32), int64(math.MaxInt32)
	for _, snap := range snaps {
		if snap.SnapID > max {
			max = snap.SnapID
		}
		if snap.SnapID < min {
			min = snap.SnapID
		}
	}
	return min, max, nil
}

func (p *PerfCollecter) queryDatabaseInstance(log yaslog.YasLog) (*yasdb.WrmDatabaseInstance, error) {
	tx := yasqlutil.GetLocalInstance(p.YasdbUser, p.YasdbPassword, p.YasdbHome, p.YasdbData)
	dataInstance, err := yasdb.QueryWrmDatabaseInstance(tx)
	if err != nil {
		log.Errorf("query wrm$database_instance err: %s", err.Error())
		return nil, err
	}
	return dataInstance, nil
}

// locally saved awr path
func (p *PerfCollecter) getAWRPath() string {
	return path.Join(_packageDir, ytccollectcommons.YASDB_DIR_NAME, AWR)
}

// locally saved slow path
func (p *PerfCollecter) getSlowPath() string {
	return path.Join(_packageDir, ytccollectcommons.YASDB_DIR_NAME, SLOW)
}

func (p *PerfCollecter) deleteSqlFile(sqlPath string) {
	if fileutil.IsExist(sqlPath) {
		_ = os.Remove(sqlPath)
	}
}

func (p *PerfCollecter) genAWRHtmlReport(log yaslog.YasLog, sqlFile string) (string, error) {
	execResult := make(chan execRes)
	timeout := confdef.GetStrategyConf().Collect.GetAWRTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go p.genAWRReport(log, sqlFile, execResult)
	select {
	case <-ctx.Done():
		err := errors.New("gen awr report timeout")
		log.Error(err)
		return "", err
	case res := <-execResult:
		ret := res.ret
		stdout := res.stdout
		stderr := res.stderr
		if ret != 0 {
			return "", fmt.Errorf("execute sql error, create awr report err: %s", stderr)
		}
		if strings.Contains(stdout, YASQL_ERR_PREFIX) {
			return "", fmt.Errorf("execute sql error, err: %s", stdout)
		}
		htmlFile, err := p.createAWRHtmlReport(log, stdout)
		if err != nil {
			log.Errorf("create awr report err: %s", err.Error())
			return "", err
		}
		return htmlFile, nil
	}
}

func (p *PerfCollecter) genAWRReport(log yaslog.YasLog, sqlFile string, res chan execRes) {
	yasqlBin := path.Join(p.YasdbHome, yasqlutil.BIN_PATH, yasqlutil.YASQL_BIN)
	connectPath := fmt.Sprintf("%s/%s", p.YasdbUser, p.YasdbPassword)
	env := []string{
		fmt.Sprintf("%s=%s", yasqlutil.LIB_KEY, path.Join(p.YasdbHome, yasqlutil.LIB_PATH)),
		fmt.Sprintf("%s=%s", yasqlutil.YASDB_DATA, p.YasdbData),
	}
	cmd := fmt.Sprintf("%s %s -f %s", yasqlBin, connectPath, sqlFile)
	exec := execer.NewExecer(log)
	ret, stdout, stderr := exec.EnvExec(env, bashdef.CMD_BASH, "-c", cmd)
	res <- execRes{
		ret:    ret,
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *PerfCollecter) createAWRHtmlReport(log yaslog.YasLog, stdout string) (string, error) {
	awrPath := p.getAWRPath()
	startStr, endStr := p.genStartEndStr(timedef.TIME_FORMAT_IN_FILE)
	htmlFile := path.Join(awrPath, fmt.Sprintf("awrrpt-%s-%s.html", startStr, endStr))
	if err := fileutil.WriteFile(htmlFile, []byte(stdout)); err != nil {
		return "", err
	}
	return htmlFile, nil
}

func (p *PerfCollecter) collectSlowSQL() error {
	log := log.Module.M(datadef.PERF_YASDB_SLOW_SQL)
	slowSQL := &datadef.YTCItem{
		Name:     datadef.PERF_YASDB_SLOW_SQL,
		Children: make(map[string]datadef.YTCItem),
	}
	defer p.fillResult(slowSQL)
	slowSQL.Children[KEY_SLOW_SQL_LOGS_IN_FILE] = *p.collectSlowLogsInFile(log)
	if p.yasdbValidateErr == nil {
		slowSQL.Children[KEY_SLOW_SQL_LOGS_IN_TABLE] = *p.collectSlowLogsInTable(log)
		slowSQL.Children[KEY_SLOW_SQL_PARAMETER] = *p.collectSlowParameter(log)
	}
	return nil
}

func (p *PerfCollecter) collectSlowLogsInTable(log yaslog.YasLog) (slowLogs *datadef.YTCItem) {
	slowLogs = new(datadef.YTCItem)
	slows, err := p.querySlowSql(log)
	if err != nil {
		slowLogs.Error = err.Error()
		log.Errorf("query slow log err: %s", err.Error())
		slowLogs.Description = datadef.GenGetDatabaseViewDesc(VIEW_SLOW_LOG)
		return
	}
	slowLogs.Details = slows
	return slowLogs
}

func (p *PerfCollecter) collectSlowParameter(log yaslog.YasLog) (parameter *datadef.YTCItem) {
	parameter = new(datadef.YTCItem)
	res := make([]*yasdb.VParameter, 0)
	for _, key := range _slowParameter {
		tx := yasqlutil.GetLocalInstance(p.YasdbUser, p.YasdbPassword, p.YasdbHome, p.YasdbData)
		value, err := yasdb.QueryParameter(tx, key)
		if err != nil {
			parameter.Error = err.Error()
			parameter.Description = datadef.GenGetDatabaseParameterDesc(string(key))
			log.Errorf("get slow parameter: %s err: %s", key, err.Error())
			return
		}
		res = append(res, &yasdb.VParameter{
			Name:  string(key),
			Value: value,
		})
	}
	parameter.Details = res
	return
}

func (p *PerfCollecter) collectSlowLogsInFile(log yaslog.YasLog) (cutSlowLog *datadef.YTCItem) {
	cutSlowLog = new(datadef.YTCItem)
	slowPath, err := p.saveSlowLog(log)
	if err != nil {
		log.Errorf("get slow log err: %s", err.Error())
		cutSlowLog.Error = err.Error()
		cutSlowLog.Description = datadef.GenDefaultDesc()
		return
	}
	relative, err := filepath.Rel(_packageDir, slowPath)
	if err != nil {
		log.Errorf("get relative path err: %s", err.Error())
		cutSlowLog.Error = err.Error()
		cutSlowLog.Description = datadef.GenDefaultDesc()
		return
	}
	cutSlowLog.Details = p.GenPackageRelativePath(relative)
	return
}

func (p *PerfCollecter) saveSlowLog(log yaslog.YasLog) (string, error) {
	logLines, err := p.querySlowSqlFromFile(log)
	if err != nil {
		return "", err
	}
	slowPath := p.getSlowPath()
	slowFile := path.Join(slowPath, ytccollectcommons.SLOW_LOG)
	if err := fileutil.WriteFile(slowFile, []byte(strings.Join(logLines, stringutil.STR_NEWLINE)+stringutil.STR_NEWLINE)); err != nil {
		return "", err
	}
	return slowFile, nil
}

func (p *PerfCollecter) querySlowSqlFromFile(log yaslog.YasLog) ([]string, error) {
	var slowPath string
	var err error
	if p.yasdbValidateErr == nil {
		slowPath, err = p.getSlowLogPath()
		if err != nil {
			log.Errorf("get slow log path err: %s", err.Error())
			return nil, err
		}
	} else {
		slowPath = path.Join(p.YasdbData, ytccollectcommons.LOG, ytccollectcommons.SLOW)
	}
	slowLog := path.Join(slowPath, ytccollectcommons.SLOW_LOG)
	return p.filterSlowLog(slowLog, log)
}

func (p *PerfCollecter) filterSlowLog(slowLog string, log yaslog.YasLog) ([]string, error) {
	slowLogFn, err := os.Open(slowLog)
	if err != nil {
		log.Errorf("open slow log err :%s", err.Error())
		return nil, err
	}
	scanner := bufio.NewScanner(slowLogFn)
	var lines []string
	var isCollected bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, TimePrefix) {
			timeStr := strings.TrimPrefix(line, TimePrefix)
			currentSqlTime, err := time.ParseInLocation(timedef.TIME_FORMAT, timeStr, time.Local)
			if err != nil {
				log.Errorf("parse time err: %s", err.Error())
				continue
			}
			isCollected = currentSqlTime.After(p.StartTime) && currentSqlTime.Before(p.EndTime)
		}
		if !isCollected {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func (p *PerfCollecter) querySlowSql(log yaslog.YasLog) ([]*yasdb.SlowLog, error) {
	tx := yasqlutil.GetLocalInstance(p.YasdbUser, p.YasdbPassword, p.YasdbHome, p.YasdbData)
	startStr, endStr := p.genStartEndStr(timedef.TIME_FORMAT)
	slows, err := yasdb.QuerySlowLog(tx, startStr, endStr)
	if err != nil {
		log.Errorf("get slow log err: %s", err.Error())
		return nil, err
	}
	return slows, nil
}

func (p *PerfCollecter) genStartEndStr(layout string) (string, string) {
	return p.StartTime.Format(layout), p.EndTime.Format(layout)
}
