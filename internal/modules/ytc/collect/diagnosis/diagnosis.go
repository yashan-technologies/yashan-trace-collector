package diagnosis

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	"ytc/defs/errdef"
	"ytc/defs/timedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/fileutil"
	"ytc/utils/processutil"
	"ytc/utils/stringutil"
	"ytc/utils/timeutil"
	"ytc/utils/userutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/fs"
)

type checkFunc func() *ytccollectcommons.NoAccessRes

const (
	CORE_PATTERN_PATH = "/proc/sys/kernel/core_pattern"

	ABRT_HOOK_CPP         = "abrt-hook-ccpp"
	ABRT_CONF             = "/etc/abrt/abrt.conf"
	DUMP_LOCATION         = "DumpLocation"
	DEFAULT_DUMP_LOCATION = "/var/spool/abrt"

	SYSTEMD_COREDUMP         = "systemd-coredump"
	SYSTEMD_COREDUMP_CONF    = "/etc/systemd/coredump.conf"
	STORAGE                  = "Storage"
	STORAGE_EXTERNAL         = "external"
	EXTERNAL_STORAGE         = "ExternalStorage"
	DEFAULT_EXTERNAL_STORAGE = "/var/lib/systemd/coredump"
	LOG_ROTATE_CONFIG        = "/etc/logrotate.conf"

	SYSTEM_LOG_MESSAGES = "/var/log/messages"
	SYSTEM_LOG_SYSLOG   = "/var/log/syslog"

	DIAG_DIR_NAME  = "diag"
	LOG_DIR_NAME   = "log"
	EXTRA_DIR_NAME = "extra"

	CORE_DUMP_DIR_NAME = "coredump"

	YASDB_DIR_NAME  = "yasdb"
	SYSTEM_DIR_NAME = "system"

	YASDB_ALERT_LOG = "alert"
	YASDB_RUN_LOG   = "run"

	SYSTEM_DMESG_LOG    = "dmesg"
	SYSTEM_MESSAGES_LOG = "messages"
	SYSTEM_SYS_LOG      = "syslog"

	LOG_FILE_SUFFIX = "%s.log"
	TAR_FILE_SUFFIX = "%s.tar.gz"

	CORE_FILE_KEY = "core"
)

const (
	_getErrMessage = "get\t[%s]\terror:\t%s"
)

var (
	DiagChineseName = map[string]string{
		datadef.DIAG_YASDB_ADR:             "数据库ADR日志",
		datadef.DIAG_YASDB_RUNLOG:          "数据库run.log日志",
		datadef.DIAG_YASDB_ALERTLOG:        "数据库alert.log日志",
		datadef.DIAG_YASDB_PROCESS_STATUS:  "数据库进程信息",
		datadef.DIAG_YASDB_INSTANCE_STATUS: "数据库实例状态",
		datadef.DIAG_YASDB_DATABASE_STATUS: "数据库状态",
		datadef.DIAG_HOST_SYSTEMLOG:        "操作系统日志",
		datadef.DIAG_HOST_KERNELLOG:        "操作系统内核日志",
		datadef.DIAG_YASDB_COREDUMP:        "Core Dump",
	}
)

var _packageDir = ""

type logTimeParseFunc func(date time.Time, line string) (time.Time, error)

type DiagCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *datadef.YTCModule
	yasdbValidateErr error
	notConnectDB     bool
}

func NewDiagCollecter(collectParam *collecttypedef.CollectParam) *DiagCollecter {
	return &DiagCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &datadef.YTCModule{
			Module: collecttypedef.TYPE_DIAG,
		},
	}
}

func (d *DiagCollecter) CheckAccess(yasdbValidate error) (noAccess []ytccollectcommons.NoAccessRes) {
	d.yasdbValidateErr = yasdbValidate
	noAccess = make([]ytccollectcommons.NoAccessRes, 0)
	funcMap := d.CheckFunc()
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
func (d *DiagCollecter) CollectFunc(items []string) (res map[string]func() error) {
	res = make(map[string]func() error)
	itemFuncMap := d.itemFunc()
	for _, collectItem := range items {
		_, ok := itemFuncMap[collectItem]
		if !ok {
			log.Module.Errorf("get %s collect func err %s", collectItem)
			continue
		}
		res[collectItem] = itemFuncMap[collectItem]
	}
	return
}

// [Interface Func]
func (b *DiagCollecter) Type() string {
	return collecttypedef.TYPE_DIAG
}

// [Interface Func]
func (b *DiagCollecter) CollectedItem(noAccess []ytccollectcommons.NoAccessRes) (res []string) {
	noMap := b.getNotAccessItem(noAccess)
	for item := range DiagChineseName {
		if _, ok := noMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

func (b *DiagCollecter) getNotAccessItem(noAccess []ytccollectcommons.NoAccessRes) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, noAccessRes := range noAccess {
		if noAccessRes.ForceCollect {
			continue
		}
		res[noAccessRes.ModuleItem] = struct{}{}
	}
	return
}

// [Interface Func]
func (b *DiagCollecter) Start(packageDir string) (err error) {
	b.setPackageDir(packageDir)
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, CORE_DUMP_DIR_NAME)); err != nil {
		return
	}
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)); err != nil {
		return
	}
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)); err != nil {
		return
	}
	return
}

func (b *DiagCollecter) setPackageDir(packageDir string) {
	_packageDir = packageDir
	log.Module.Infof("package dir is %s", _packageDir)
}

// [Interface Func]
func (b *DiagCollecter) Finish() *datadef.YTCModule {
	return b.ModuleCollectRes
}

func (b *DiagCollecter) fillResult(data *datadef.YTCItem) {
	b.ModuleCollectRes.Set(data)
}

func (b *DiagCollecter) yasdbProcessStatus() (err error) {
	yasdbProcessStatusItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_PROCESS_STATUS}
	defer b.fillResult(&yasdbProcessStatusItem)

	log := log.Module.M(datadef.DIAG_YASDB_PROCESS_STATUS)
	processes, err := processutil.GetYasdbProcess(b.YasdbData)
	if err != nil {
		log.Error(err)
		yasdbProcessStatusItem.Error = err.Error()
		return
	}
	if len(processes) == 0 {
		err = processutil.ErrYasdbProcessNotFound
		log.Error(err)
		yasdbProcessStatusItem.Error = err.Error()
		return
	}
	proc := processes[0]
	if err = proc.FindBaseInfo(); err != nil {
		log.Error(err)
		yasdbProcessStatusItem.Error = err.Error()
		return
	}
	yasdbProcessStatusItem.Details = proc
	return
}

func (b *DiagCollecter) yasdbInstanceStatus() (err error) {
	yasdbInstanceStatusItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_INSTANCE_STATUS}
	defer b.fillResult(&yasdbInstanceStatusItem)

	log := log.Module.M(datadef.DIAG_YASDB_INSTANCE_STATUS)
	if b.notConnectDB {
		err = fmt.Errorf("connect failed, skip")
		yasdbInstanceStatusItem.Error = err.Error()
		log.Error(err)
		return
	}
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	data, err := yasdb.QueryInstance(tx)
	if err != nil {
		log.Error(err)
		yasdbInstanceStatusItem.Error = err.Error()
		return
	}
	yasdbInstanceStatusItem.Details = data
	return
}

func (b *DiagCollecter) yasdbDatabaseStatus() (err error) {
	yasdbDatabaseStatusItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_DATABASE_STATUS}
	defer b.fillResult(&yasdbDatabaseStatusItem)

	log := log.Module.M(datadef.DIAG_YASDB_DATABASE_STATUS)
	if b.notConnectDB {
		err = fmt.Errorf("connect failed, skip")
		yasdbDatabaseStatusItem.Error = err.Error()
		log.Error(err)
		return
	}
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	data, err := yasdb.QueryDatabase(tx)
	if err != nil {
		log.Error(err)
		yasdbDatabaseStatusItem.Error = err.Error()
		return
	}
	yasdbDatabaseStatusItem.Details = data
	return
}

func (b *DiagCollecter) yasdbADRLog() (err error) {
	yasdbADRLogItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_ADR}
	defer b.fillResult(&yasdbADRLogItem)

	log := log.Module.M(datadef.DIAG_YASDB_ADR)
	adrPath := path.Join(b.YasdbData, DIAG_DIR_NAME) // default adr log path
	if !b.notConnectDB {
		if adrPath, err = GetAdrPath(b.CollectParam); err != nil {
			log.Error(err)
			yasdbADRLogItem.Error = err.Error()
			return
		}
	}
	if !fs.IsDirExist(adrPath) {
		err = &errdef.ErrFileNotFound{Fname: adrPath}
		log.Error(err)
		yasdbADRLogItem.Error = err.Error()
		return
	}
	// package adr to dest
	destPath := path.Join(_packageDir, DIAG_DIR_NAME)
	destFile := fmt.Sprintf("yasdb-diag-%s.tar.gz", time.Now().Format(timedef.TIME_FORMAT_IN_FILE))
	// 这个函数只会将非空的文件夹下的内容打包出来，如果文件夹是空的，不会在目标压缩包中创建文件夹
	if err = fs.TarDir(adrPath, path.Join(destPath, destFile)); err != nil {
		log.Error(err)
		yasdbADRLogItem.Error = err.Error()
		return
	}
	yasdbADRLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, destFile))
	return
}

func (b *DiagCollecter) yasdbCoredumpFile() (err error) {
	yasdbCoreDumpItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_COREDUMP}
	defer b.fillResult(&yasdbCoreDumpItem)

	log := log.Module.M(datadef.DIAG_YASDB_COREDUMP)
	coreDumpPath, err := GetCoredumpPath()
	if err != nil {
		log.Error(err)
		yasdbCoreDumpItem.Error = err.Error()
		return
	}
	if !path.IsAbs(coreDumpPath) {
		coreDumpPath = path.Join(b.YasdbHome, "bin", coreDumpPath)
	}
	log.Infof("core dump file path is: %s", coreDumpPath)
	coreFileKey := confdef.GetStrategyConf().Collect.CoreFileKey
	if stringutil.IsEmpty(coreFileKey) {
		coreFileKey = CORE_FILE_KEY
	}
	files, err := os.ReadDir(coreDumpPath)
	if err != nil {
		log.Error(err)
		yasdbCoreDumpItem.Error = err.Error()
		return
	}
	for _, file := range files {
		if !file.Type().IsRegular() || !strings.Contains(file.Name(), coreFileKey) {
			continue
		}
		info, e := file.Info()
		if e != nil {
			err = e
			log.Error(err)
			yasdbCoreDumpItem.Error = err.Error()
			return
		}
		createAt := info.ModTime()
		if createAt.Before(b.StartTime) || createAt.After(b.EndTime) {
			continue
		}
		if err = fs.CopyFile(path.Join(coreDumpPath, file.Name()), path.Join(_packageDir, DIAG_DIR_NAME, CORE_DUMP_DIR_NAME, file.Name())); err != nil {
			log.Error(err)
			yasdbCoreDumpItem.Error = err.Error()
			return
		}
	}
	yasdbCoreDumpItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, CORE_DUMP_DIR_NAME))
	return
}

func (b *DiagCollecter) yasdbRunLog() (err error) {
	yasdbRunLogItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_RUNLOG}
	defer b.fillResult(&yasdbRunLogItem)

	log := log.Module.M(datadef.DIAG_YASDB_RUNLOG)
	log.Debug("start to collect yasdb run.log")
	runLogPath, runLogFile := path.Join(b.YasdbData, LOG_DIR_NAME, YASDB_RUN_LOG), fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_RUN_LOG)
	if !b.notConnectDB {
		if runLogPath, err = GetYasdbRunLogPath(b.CollectParam); err != nil {
			log.Error(err)
			yasdbRunLogItem.Error = err.Error()
			return
		}
	}
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)
	// get run log files
	runLogFiles, err := b.getLogFiles(log, runLogPath, YASDB_RUN_LOG)
	if err != nil {
		log.Error(err)
		yasdbRunLogItem.Error = err.Error()
		return
	}
	// write run log to dest
	if err = b.collectYasdbRunLog(log, runLogFiles, path.Join(destPath, runLogFile), b.StartTime, b.EndTime); err != nil {
		log.Error(err)
		yasdbRunLogItem.Error = err.Error()
		return
	}
	yasdbRunLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME, runLogFile))
	return
}

func (b *DiagCollecter) collectYasdbRunLog(log yaslog.YasLog, srcs []string, dest string, start, end time.Time) (err error) {
	timeParseFunc := func(date time.Time, line string) (t time.Time, err error) {
		fields := strings.Split(line, stringutil.STR_BLANK_SPACE)
		if len(fields) < 2 {
			err = fmt.Errorf("invalid line: %s, skip", line)
			return
		}
		timeStr := fmt.Sprintf("%s %s", fields[0], fields[1])
		return time.ParseInLocation(timedef.TIME_FORMAT_WITH_MICROSECOND, timeStr, time.Local)
	}
	for _, f := range srcs {
		logEndTime := time.Now()
		if path.Base(f) != fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_RUN_LOG) {
			fileds := strings.Split(strings.TrimSuffix(path.Base(f), ".log"), stringutil.STR_HYPHEN)
			if len(fileds) < 2 {
				log.Errorf("failed to get log end time from %s, skip", f)
				continue
			}
			if logEndTime, err = time.ParseInLocation(timedef.TIME_FORMAT_IN_FILE, fileds[1], time.Local); err != nil {
				log.Errorf("failed to parse log end time from %s", fileds[1])
				continue
			}
		}
		if logEndTime.Before(b.StartTime) { // no need to write into dest
			log.Debugf("skip run log file: %s", f)
			continue
		}
		if err = b.collectLog(log, f, dest, time.Now(), timeParseFunc); err != nil {
			return
		}
	}
	return
}

func (b *DiagCollecter) yasdbAlertLog() (err error) {
	yasdbAlertLogItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_ALERTLOG}
	defer b.fillResult(&yasdbAlertLogItem)

	log := log.Module.M(datadef.DIAG_YASDB_ALERTLOG)
	logPath := path.Join(b.YasdbData, LOG_DIR_NAME)
	alertLogPath, alertLogFile := path.Join(logPath, YASDB_ALERT_LOG), fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_ALERT_LOG)
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)
	// get alert log
	timeParseFunc := func(date time.Time, line string) (t time.Time, e error) {
		fields := strings.Split(line, stringutil.STR_BAR)
		return time.ParseInLocation(timedef.TIME_FORMAT_WITH_MICROSECOND, fields[0], time.Local)
	}
	srcFile, destFile := path.Join(alertLogPath, alertLogFile), path.Join(destPath, alertLogFile)
	if err = b.collectLog(log, srcFile, destFile, time.Now(), timeParseFunc); err != nil {
		log.Error(err)
		yasdbAlertLogItem.Error = err.Error()
		return
	}
	yasdbAlertLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME, alertLogFile))
	return
}

func (b *DiagCollecter) hostKernelLog() (err error) {
	hostKernelLogItem := datadef.YTCItem{Name: datadef.DIAG_HOST_KERNELLOG}
	defer b.fillResult(&hostKernelLogItem)

	log := log.Module.M(datadef.DIAG_HOST_KERNELLOG)
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)
	// dmesg.log
	execer := execerutil.NewExecer(log)
	dmesgFile := fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_DMESG_LOG)
	dest := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_DMESG_LOG))
	ret, stdout, stderr := execer.Exec(bashdef.CMD_BASH, "-c", bashdef.CMD_DMESG)
	if ret != 0 {
		err = fmt.Errorf("failed to get host dmesg log, err: %s", stderr)
		log.Error(err)
		hostKernelLogItem.Error = err.Error()
		return
	}
	// write to dest
	if err = fileutil.WriteFile(dest, []byte(stdout)); err != nil {
		log.Error(err)
		hostKernelLogItem.Error = err.Error()
		return
	}
	hostKernelLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME, dmesgFile))
	return
}

func (b *DiagCollecter) hostSystemLog() (err error) {
	hostSystemLogItem := datadef.YTCItem{
		Name:     datadef.DIAG_HOST_SYSTEMLOG,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostSystemLogItem)

	log := log.Module.M(datadef.DIAG_HOST_SYSTEMLOG)
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)
	if userutil.IsCurrentUserRoot() {
		// message.log
		destMessageLogFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_MESSAGES_LOG))
		if err = b.collectHostLog(log, SYSTEM_LOG_MESSAGES, destMessageLogFile, SYSTEM_MESSAGES_LOG); err != nil {
			log.Error(err)
			hostSystemLogItem.Children[SYSTEM_MESSAGES_LOG] = datadef.YTCItem{Error: err.Error()}
		} else {
			logPath := fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_MESSAGES_LOG)))
			hostSystemLogItem.Children[SYSTEM_MESSAGES_LOG] = datadef.YTCItem{Details: logPath}
		}
		// syslog.log
		destSysLogFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_SYS_LOG))
		if err = b.collectHostLog(log, SYSTEM_LOG_SYSLOG, destSysLogFile, SYSTEM_SYS_LOG); err != nil {
			log.Error(err)
			hostSystemLogItem.Children[SYSTEM_SYS_LOG] = datadef.YTCItem{Error: err.Error()}
		} else {
			logPath := fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_SYS_LOG)))
			hostSystemLogItem.Children[SYSTEM_SYS_LOG] = datadef.YTCItem{Details: logPath}
		}
	} else {
		message := "has no permission to collect system log"
		description := "没有权限收集系统日志"
		hostSystemLogItem.Children[SYSTEM_MESSAGES_LOG] = datadef.YTCItem{
			Error:       message,
			Description: description,
		}
		hostSystemLogItem.Children[SYSTEM_SYS_LOG] = datadef.YTCItem{
			Error:       message,
			Description: description,
		}
	}
	return
}

func (b *DiagCollecter) collectHostLog(log yaslog.YasLog, src, dest string, prefix string) (err error) {
	hasSetDateext, err := b.hasSetDateext()
	if err != nil {
		return
	}
	if hasSetDateext {
		return b.collectHostLogWithSetDateext(log, src, dest, prefix)
	}
	return b.collectHostLogWithoutSetDateext(log, src, dest)
}

func (b *DiagCollecter) hostLogTimeParse(date time.Time, line string) (t time.Time, err error) {
	fields := strings.Split(line, stringutil.STR_BLANK_SPACE)
	if len(fields) < 3 {
		err = fmt.Errorf("invalid line: %s, skip", line)
		return
	}
	tmpTime, err := time.ParseInLocation(timedef.TIME_FORMAT_TIME, fields[2], time.Local)
	if err != nil {
		return
	}
	hour, min, sec := tmpTime.Hour(), tmpTime.Minute(), tmpTime.Second()
	day, err := strconv.Atoi(fields[1])
	if err != nil {
		return
	}
	mon, err := timeutil.GetMonth(fields[0])
	year := date.Year()
	if date.Month() < mon {
		year = year - 1
	}
	t = time.Date(year, mon, day, hour, min, sec, 0, time.Local)
	return
}

func (b *DiagCollecter) collectHostLogWithSetDateext(log yaslog.YasLog, src, dest string, prefix string) (err error) {
	var srcs []string
	srcs, err = b.getLogFiles(log, path.Dir(src), prefix)
	if err != nil {
		return
	}
	var logFiles []string // resort logFile so that the current log file is the last one, other file sorted by time is in the first
	for _, v := range srcs {
		if v == src {
			continue
		}
		logFiles = append(logFiles, v)
	}
	if len(srcs) != len(logFiles) {
		logFiles = append(logFiles, src)
	}
	for _, logFile := range logFiles {
		log.Debugf("try to collect %s", logFile)
		date := time.Now()
		if logFile != src {
			fileds := strings.Split(path.Base(logFile), stringutil.STR_HYPHEN)
			if len(fileds) < 2 {
				log.Errorf("failed to get log end time from %s, skip", logFile)
				continue
			}
			// get date from log file name
			date, err = time.ParseInLocation(timedef.TIME_FORMAT_DATE_IN_FILE, fileds[1], time.Local)
			if err != nil {
				log.Error("failed to get date from: %s, err: %s", logFile, err.Error())
				continue
			}
			// try to get log end time from last 3 line in log
			k := 3
			lastKLines, err := fileutil.Tail(logFile, k)
			if err != nil {
				log.Errorf("failed to read file %s last %d line, err: %s", logFile, k, err.Error())
			} else {
				for i := 0; i < len(lastKLines); i++ {
					if stringutil.IsEmpty(lastKLines[i]) {
						continue
					}
					var tmpData time.Time
					tmpData, err = b.hostLogTimeParse(date, stringutil.RemoveExtraSpaces(strings.TrimSpace(lastKLines[i])))
					if err != nil {
						log.Errorf("failed to parse time from line: %s, err: %s", lastKLines[i], err.Error())
						continue
					}
					date = tmpData
				}
			}
			log.Debugf("log file %s end date is %s", logFile, date)
			if date.Before(b.StartTime) {
				log.Infof("skip to collect log %s, log file end date: %s , collect start date %s", logFile, date.AddDate(0, 0, -1), b.StartTime)
				continue
			}
		}
		if err = b.collectLog(log, logFile, dest, date, b.hostLogTimeParse); err != nil {
			log.Errorf("failed to collect from: %s, err: %s", logFile, err.Error())
			continue
		}
		log.Debugf("succeed to collect %s", logFile)
	}
	return
}

func (b *DiagCollecter) collectHostLogWithoutSetDateext(log yaslog.YasLog, src, dest string) (err error) {
	// get log file last modify time
	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}
	srcModTime := srcInfo.ModTime()
	if srcModTime.Before(b.StartTime) {
		log.Infof("log %s last modify time is %s, skip", src, srcModTime)
		return
	}
	return b.reverseCollectLog(log, src, dest, srcModTime, b.hostLogTimeParse)
}

func (b *DiagCollecter) hasSetDateext() (res bool, err error) {
	config, err := os.Open(LOG_ROTATE_CONFIG)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(config)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "dateext") {
			res = true
			return
		}
	}
	return
}

func (b *DiagCollecter) getLogFiles(log yaslog.YasLog, logPath string, prefix string) (logFiles []string, err error) {
	entrys, err := os.ReadDir(logPath)
	if err != nil {
		log.Error(err)
		return
	}
	for _, entry := range entrys {
		if !entry.Type().IsRegular() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		logFiles = append(logFiles, path.Join(logPath, entry.Name()))
	}
	// sort with file name
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i] < logFiles[j]
	})
	return
}

// some log may not contain date info in the log file content, but in the log name
func (b *DiagCollecter) collectLog(log yaslog.YasLog, src, dest string, date time.Time, timeParseFunc logTimeParseFunc) (err error) {
	destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileutil.DEFAULT_FILE_MODE)
	if err != nil {
		return
	}
	defer destFile.Close()
	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer srcFile.Close()

	var t time.Time
	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		txt := scanner.Text()
		line := stringutil.RemoveExtraSpaces(strings.TrimSpace(txt))
		if stringutil.IsEmpty(line) {
			continue
		}
		if t, err = timeParseFunc(date, line); err != nil {
			log.Error("skip line: %s, err: %s", txt, err.Error())
			continue
		}
		if t.Before(b.StartTime) {
			continue
		}
		if t.After(b.EndTime) {
			break
		}
		_, err = destFile.WriteString(txt + stringutil.STR_NEWLINE)
		if err != nil {
			return
		}
	}
	log.Debugf("succeed to write log file %s to %s", src, dest)
	return
}

func (b *DiagCollecter) reverseCollectLog(log yaslog.YasLog, src, dest string, date time.Time, timeParseFunc logTimeParseFunc) (err error) {
	// open tmp file
	tmp := fmt.Sprintf("%s.temp", dest)
	tmpFile, err := os.OpenFile(tmp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileutil.DEFAULT_FILE_MODE)
	if err != nil {
		return
	}
	defer os.Remove(tmp)
	// open src file in reverse order
	reverseSrcFile, err := fileutil.NewReverseFile(src)
	if err != nil {
		return
	}
	defer reverseSrcFile.Close()
	for {
		line, e := reverseSrcFile.ReadLine()
		if e != nil {
			if e == io.EOF { // read to end
				break
			}
			err = e
			return
		}
		var t time.Time
		t, err = timeParseFunc(date, stringutil.RemoveExtraSpaces(strings.TrimSpace(line)))
		if err != nil {
			return
		}
		if t.After(b.EndTime) {
			continue
		}
		if t.Before(b.StartTime) {
			break
		}
		// write to tmp file
		if _, err = tmpFile.WriteString(line + stringutil.STR_NEWLINE); err != nil {
			return
		}
	}
	// reverse open tmp file
	reverseTmpFile, err := fileutil.NewReverseFile(tmp)
	if err != nil {
		return
	}
	defer reverseTmpFile.Close()
	// open dest file
	destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileutil.DEFAULT_FILE_MODE)
	if err != nil {
		return
	}
	defer destFile.Close()
	for {
		line, e := reverseTmpFile.ReadLine()
		if e != nil {
			if e == io.EOF { // read to end
				break
			}
			err = e
			return
		}
		// write to dest file
		if _, err = destFile.WriteString(line + stringutil.STR_NEWLINE); err != nil {
			return
		}
	}
	log.Debugf("succeed to write log file %s to %s", src, dest)
	return
}

func (b *DiagCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		datadef.DIAG_YASDB_PROCESS_STATUS:  b.yasdbProcessStatus,
		datadef.DIAG_YASDB_INSTANCE_STATUS: b.yasdbInstanceStatus,
		datadef.DIAG_YASDB_DATABASE_STATUS: b.yasdbDatabaseStatus,
		datadef.DIAG_YASDB_ADR:             b.yasdbADRLog,
		datadef.DIAG_YASDB_ALERTLOG:        b.yasdbAlertLog,
		datadef.DIAG_YASDB_RUNLOG:          b.yasdbRunLog,
		datadef.DIAG_YASDB_COREDUMP:        b.yasdbCoredumpFile,
		datadef.DIAG_HOST_SYSTEMLOG:        b.hostSystemLog,
		datadef.DIAG_HOST_KERNELLOG:        b.hostKernelLog,
	}
}
