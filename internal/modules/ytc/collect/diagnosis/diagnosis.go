package diagnosis

import (
	"fmt"
	"os"

	"path"
	"strings"
	"time"
	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	"ytc/defs/errdef"
	"ytc/defs/timedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/userutil"

	"ytc/utils/processutil"
	"ytc/utils/stringutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yasutil/execer"
	"git.yasdb.com/go/yasutil/fs"
)

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

	SYSTEM_LOG_MESSAGES = "/var/log/messages"
	SYSTEM_LOG_SYSLOG   = "/var/log/syslog"

	DIAG_DIR_NAME = "diag"
	LOG_DIR_NAME  = "log"

	CORE_DUMP_DIR_NAME = "coredump"

	YASDB_DIR_NAME  = "yasdb"
	SYSTEM_DIR_NAME = "system"

	YASDB_ALERT_LOG = "alert"
	YASDB_RUN_LOG   = "run"

	SYSTEM_DMESG_LOG   = "dmesg"
	SYSTEM_MESSAGE_LOG = "message"
	SYSTEM_SYS_LOG     = "syslog"

	LOG_FILE_SUFFIX = "%s.log"
	TAR_FILE_SUFFIX = "%s.tar.gz"

	CORE_FILE_KEY = "core"
)

const (
	_getErrMessage = "get\t[%s]\terror:\t%s"
)

const (
	_please_run_with_sudo               = "you can run 'sudo ytcctl collect'"
	_please_run_with_root               = "you can run 'ytcctl collect' with root"
	_please_run_with_yasdb_user_or_sudo = "you can run 'ytcctl collect' with yasdb user or root"
)

var (
	diag_default_ch_map = map[string]string{
		data.DIAG_YASDB_ADR:             "数据库ADR日志",
		data.DIAG_YASDB_RUNLOG:          "数据库run.log日志",
		data.DIAG_YASDB_ALERTLOG:        "数据库alert.log日志",
		data.DIAG_YASDB_PROCESS_STATUS:  "数据库进程信息",
		data.DIAG_YASDB_INSTANCE_STATUS: "数据库实例状态",
		data.DIAG_YASDB_DATABASE_STATUS: "数据库状态",
		data.DIAG_HOST_SYSTEMLOG:        "操作系统日志",
		data.DIAG_YASDB_COREDUMP:        "Core Dump",
	}
)

var _package_dir = ""

type DiagCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *data.YtcModule
	notConnectDB     bool
}

func NewDiagCollecter(collectParam *collecttypedef.CollectParam) *DiagCollecter {
	return &DiagCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &data.YtcModule{
			Module: collecttypedef.TYPE_DIAG,
		},
	}
}

func (d *DiagCollecter) CheckAccess() (noAccess []data.NoAccessRes) {
	noAccess = make([]data.NoAccessRes, 0)
	itemPath := GetDiagPath(d.CollectParam)
	for item, path := range itemPath {
		if err := fileutil.CheckAccess(path); err != nil {
			var (
				desc string
				tips string
			)
			user, userErr := userutil.GetCurrentUser()
			if userErr != nil {
				log.Module.Errorf("get current user err: %s", userErr.Error())
			}
			desc = fmt.Sprintf("current user: %s %s", user, err.Error())
			switch item {
			case data.DIAG_YASDB_COREDUMP, data.DIAG_HOST_SYSTEMLOG:
				if err := userutil.CheckSudovn(log.Module); err != nil {
					if err == userutil.ErrSudoNeedPwd {
						tips = _please_run_with_sudo
					} else {
						tips = _please_run_with_root
					}
				}
			case data.DIAG_YASDB_ADR, data.DIAG_YASDB_ALERTLOG, data.DIAG_YASDB_RUNLOG:
				tips = _please_run_with_yasdb_user_or_sudo
			}
			noAccess = append(noAccess, data.NoAccessRes{
				ModuleItem:  item,
				Description: desc,
				Tips:        tips,
			})
		}
	}
	if err := d.checkYasdbEnv(); err != nil {
		d.notConnectDB = true
		tipsFormat := "defult to %s collect"
		alert := data.NoAccessRes{
			ModuleItem:   data.DIAG_YASDB_ALERTLOG,
			Description:  "alert.log may not be collected",
			Tips:         fmt.Sprintf(tipsFormat, path.Join(d.YasdbData, "alert", "alert.log")),
			ForceCollect: true,
		}
		run := data.NoAccessRes{
			ModuleItem:   data.DIAG_YASDB_RUNLOG,
			Description:  "run.log may not be collected",
			Tips:         fmt.Sprintf(tipsFormat, path.Join(d.YasdbData, "run", "run.log")),
			ForceCollect: true,
		}
		adr := data.NoAccessRes{
			ModuleItem:   data.DIAG_YASDB_ADR,
			Description:  "adr log may not be collected",
			Tips:         fmt.Sprintf(tipsFormat, path.Join(d.YasdbData, "diag")),
			ForceCollect: true,
		}
		tips := ytccollectcommons.GenYasdbEnvErrTips(err)
		instanceStatus := data.NoAccessRes{
			ModuleItem:  data.DIAG_YASDB_INSTANCE_STATUS,
			Description: err.Error(),
			Tips:        tips,
		}
		databaseStatus := data.NoAccessRes{
			ModuleItem:  data.DIAG_YASDB_DATABASE_STATUS,
			Description: err.Error(),
			Tips:        tips,
		}
		noAccess = append(noAccess, alert, run, adr, instanceStatus, databaseStatus)
	}
	return
}

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

func (b *DiagCollecter) Type() string {
	return collecttypedef.TYPE_DIAG
}

func (b *DiagCollecter) CollectedItem(noAccess []data.NoAccessRes) (res []string) {
	noMap := b.getNotAccessItem(noAccess)
	for item := range diag_default_ch_map {
		if _, ok := noMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

func (b *DiagCollecter) getNotAccessItem(noAccess []data.NoAccessRes) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, no := range noAccess {
		if no.ForceCollect {
			continue
		}
		res[no.ModuleItem] = struct{}{}
	}
	return
}

func (b *DiagCollecter) Start(packageDir string) error {
	b.setPackageDir(packageDir)
	if err := fs.Mkdir(path.Join(_package_dir, DIAG_DIR_NAME, CORE_DUMP_DIR_NAME)); err != nil {
		return err
	}
	if err := fs.Mkdir(path.Join(_package_dir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)); err != nil {
		return err
	}
	if err := fs.Mkdir(path.Join(_package_dir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)); err != nil {
		return err
	}
	return nil
}

func (b *DiagCollecter) setPackageDir(packageDir string) {
	_package_dir = packageDir
	log.Module.Infof("package dir is %s", _package_dir)
}

func (b *DiagCollecter) Finish() *data.YtcModule {
	return b.ModuleCollectRes
}

func (b *DiagCollecter) fillResult(data *data.YtcItem) {
	b.ModuleCollectRes.Lock()
	defer b.ModuleCollectRes.Unlock()
	b.ModuleCollectRes.Items = append(b.ModuleCollectRes.Items, data)
}

func (b *DiagCollecter) yasdbProcessStatus() (e error) {
	yasdbProcessStatusItem := data.YtcItem{ItemName: data.DIAG_YASDB_PROCESS_STATUS}
	defer b.fillResult(&yasdbProcessStatusItem)

	log := log.Module.M(data.DIAG_YASDB_PROCESS_STATUS)
	processes, e := processutil.GetYasdbProcess(b.YasdbData)
	if e != nil {
		log.Error(e)
		yasdbProcessStatusItem.Err = e.Error()
		return
	}
	if len(processes) == 0 {
		e = processutil.ErrYasdbProcessNotFound
		log.Error(e)
		yasdbProcessStatusItem.Err = e.Error()
		return
	}
	proc := processes[0]
	if e = proc.FindBaseInfo(); e != nil {
		log.Error(e)
		yasdbProcessStatusItem.Err = e.Error()
		return
	}
	yasdbProcessStatusItem.Details = proc
	return
}

func (b *DiagCollecter) yasdbInstanceStatus() (e error) {
	yasdbInstanceStatusItem := data.YtcItem{ItemName: data.DIAG_YASDB_INSTANCE_STATUS}
	defer b.fillResult(&yasdbInstanceStatusItem)

	log := log.Module.M(data.DIAG_YASDB_INSTANCE_STATUS)
	if b.notConnectDB {
		e = fmt.Errorf("connect failed, skip")
		yasdbInstanceStatusItem.Err = e.Error()
		log.Error(e)
		return
	}
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	data, e := yasdb.QueryInstance(tx)
	if e != nil {
		log.Error(e)
		yasdbInstanceStatusItem.Err = e.Error()
		return
	}
	yasdbInstanceStatusItem.Details = data
	return
}

func (b *DiagCollecter) yasdbDatabaseStatus() (e error) {
	yasdbDatabaseStatusItem := data.YtcItem{ItemName: data.DIAG_YASDB_DATABASE_STATUS}
	defer b.fillResult(&yasdbDatabaseStatusItem)

	log := log.Module.M(data.DIAG_YASDB_DATABASE_STATUS)
	if b.notConnectDB {
		e = fmt.Errorf("connect failed, skip")
		yasdbDatabaseStatusItem.Err = e.Error()
		log.Error(e)
		return
	}
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	data, e := yasdb.QueryDatabase(tx)
	if e != nil {
		log.Error(e)
		yasdbDatabaseStatusItem.Err = e.Error()
		return
	}
	yasdbDatabaseStatusItem.Details = data
	return
}

func (b *DiagCollecter) yasdbADRLog() (e error) {
	yasdbADRLogItem := data.YtcItem{ItemName: data.DIAG_YASDB_ADR}
	defer b.fillResult(&yasdbADRLogItem)

	log := log.Module.M(data.DIAG_YASDB_ADR)
	adrPath := path.Join(b.YasdbData, DIAG_DIR_NAME) // default adr log path
	if !b.notConnectDB {
		if adrPath, e = GetAdrPath(b.CollectParam); e != nil {
			log.Error(e)
			yasdbADRLogItem.Err = e.Error()
			return
		}
	}
	if !fs.IsDirExist(adrPath) {
		e = &errdef.ErrFileNotFound{Fname: adrPath}
		log.Error(e)
		yasdbADRLogItem.Err = e.Error()
		return
	}
	// package adr to dest
	destPath := path.Join(_package_dir, DIAG_DIR_NAME)
	destFile := fmt.Sprintf("yasdb-diag-%s.tar.gz", time.Now().Format(timedef.TIME_FORMAT_IN_FILE))
	// TODO:这个函数只会将非空的文件夹下的内容打包出来，如果文件夹是空的，不会在目标压缩包中创建文件夹
	if e = fs.TarDir(adrPath, path.Join(destPath, destFile)); e != nil {
		log.Error(e)
		yasdbADRLogItem.Err = e.Error()
		return
	}
	yasdbADRLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, destFile))
	return
}

func (b *DiagCollecter) yasdbCoredumpFile() (e error) {
	yasdbCoreDumpItem := data.YtcItem{ItemName: data.DIAG_YASDB_COREDUMP}
	defer b.fillResult(&yasdbCoreDumpItem)

	log := log.Module.M(data.DIAG_YASDB_COREDUMP)
	coreDumpPath, e := GetCoredumpPath()
	if e != nil {
		log.Error(e)
		yasdbCoreDumpItem.Err = e.Error()
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
	files, e := os.ReadDir(coreDumpPath)
	for _, file := range files {
		if !file.Type().IsRegular() || !strings.Contains(file.Name(), coreFileKey) {
			continue
		}
		info, err := file.Info()
		if err != nil {
			log.Error(e)
			yasdbCoreDumpItem.Err = e.Error()
			return
		}
		createAt := info.ModTime()
		if createAt.Before(b.StartTime) || createAt.After(b.EndTime) {
			continue
		}
		if e = fs.CopyFile(path.Join(coreDumpPath, file.Name()), path.Join(_package_dir, DIAG_DIR_NAME, CORE_DUMP_DIR_NAME, file.Name())); e != nil {
			log.Error(e)
			yasdbCoreDumpItem.Err = e.Error()
			return
		}
	}
	yasdbCoreDumpItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, CORE_DUMP_DIR_NAME))
	return
}

func (b *DiagCollecter) yasdbRunLog() (e error) {
	yasdbRunLogItem := data.YtcItem{ItemName: data.DIAG_YASDB_RUNLOG}
	defer b.fillResult(&yasdbRunLogItem)

	log := log.Module.M(data.DIAG_YASDB_RUNLOG)
	// TODO: 暂时不实现根据日志时间处理日志的逻辑，只把run.log拿过来
	runLogPath, runLogFile := path.Join(b.YasdbData, LOG_DIR_NAME, YASDB_RUN_LOG), fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_RUN_LOG)
	if !b.notConnectDB {
		if runLogPath, e = GetYasdbRunLogPath(b.CollectParam); e != nil {
			log.Error(e)
			yasdbRunLogItem.Err = e.Error()
			return
		}
	}
	destPath := path.Join(_package_dir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)
	// get run log
	src := path.Join(runLogPath, runLogFile)
	dest := path.Join(destPath, runLogFile)
	if e = fs.CopyFile(src, dest); e != nil {
		log.Error(e)
		yasdbRunLogItem.Err = e.Error()
		return
	}
	yasdbRunLogItem.Details = fmt.Sprintf("./%s", path.Join(LOG_DIR_NAME, YASDB_DIR_NAME, runLogFile))
	return
}

func (b *DiagCollecter) yasdbAlertLog() (e error) {
	yasdbAlertLogItem := data.YtcItem{ItemName: data.DIAG_YASDB_ALERTLOG}
	defer b.fillResult(&yasdbAlertLogItem)

	// TODO: 暂时不实现根据日志时间处理日志的逻辑
	log := log.Module.M(data.DIAG_YASDB_ALERTLOG)
	logPath := path.Join(b.YasdbData, LOG_DIR_NAME)
	alertLogPath, alertLogFile := path.Join(logPath, YASDB_ALERT_LOG), fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_ALERT_LOG)
	destPath := path.Join(_package_dir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)
	if !fs.IsDirExist(destPath) {
		if e = fs.Mkdir(destPath); e != nil {
			log.Error(e)
			yasdbAlertLogItem.Err = e.Error()
			return
		}
	}
	// get alert log
	if e = fs.CopyFile(path.Join(alertLogPath, alertLogFile), path.Join(destPath, alertLogFile)); e != nil {
		log.Error(e)
		yasdbAlertLogItem.Err = e.Error()
		return
	}
	yasdbAlertLogItem.Details = fmt.Sprintf("./%s", path.Join(LOG_DIR_NAME, YASDB_DIR_NAME, alertLogFile))
	return
}

func (b *DiagCollecter) hostSystemLog() (e error) {
	hostSystemLogItem := data.YtcItem{ItemName: data.DIAG_HOST_SYSTEMLOG}
	defer b.fillResult(&hostSystemLogItem)

	log := log.Module.M(data.DIAG_HOST_SYSTEMLOG)
	var errs []string
	detailMap := make(map[string]string)
	destPath := path.Join(_package_dir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)
	if !fs.IsDirExist(destPath) {
		if e = fs.Mkdir(destPath); e != nil {
			log.Error(e)
			hostSystemLogItem.Err = e.Error()
			return
		}
	}
	// dmesg.log
	execer := execer.NewExecer(log)
	dmesgFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_DMESG_LOG))
	cmd := []string{
		"-c",
		bashdef.CMD_DMESG,
		">",
		dmesgFile,
	}
	if ret, _, stderr := execer.Exec(bashdef.CMD_BASH, cmd...); ret != 0 {
		e = fmt.Errorf("failed to get host dmesg log, err: %s", stderr)
		errs = append(errs, e.Error())
	} else {
		detailMap[SYSTEM_DMESG_LOG] = fmt.Sprintf("./%s", path.Join(LOG_DIR_NAME, dmesgFile))
	}
	if userutil.IsCurrentUserRoot() {
		// TODO:暂时只收集系统日志,不根据时间筛选
		// message.log
		messageLogFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_MESSAGE_LOG))
		if fs.IsFileExist(SYSTEM_LOG_MESSAGES) {
			if err := fs.CopyFile(SYSTEM_LOG_MESSAGES, messageLogFile); err != nil {
				errs = append(errs, err.Error())
			} else {
				detailMap[SYSTEM_MESSAGE_LOG] = fmt.Sprintf("./%s", path.Join(LOG_DIR_NAME, messageLogFile))
			}
		} else {
			errs = append(errs, fmt.Sprintf("file %s does not exist", SYSTEM_LOG_MESSAGES))
		}
		// syslog.log
		destSysLogFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_SYS_LOG))
		if fs.IsFileExist(SYSTEM_LOG_SYSLOG) {
			if err := fs.CopyFile(SYSTEM_LOG_SYSLOG, destSysLogFile); err != nil {
				errs = append(errs, err.Error())
			} else {
				detailMap[SYSTEM_SYS_LOG] = fmt.Sprintf("./%s", path.Join(LOG_DIR_NAME, messageLogFile))
			}
		} else {
			errs = append(errs, fmt.Sprintf("file %s does not exist", SYSTEM_LOG_SYSLOG))
		}
	}
	hostSystemLogItem.Details = detailMap
	hostSystemLogItem.Err = strings.Join(errs, stringutil.STR_NEWLINE)
	return
}

func (b *DiagCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		data.DIAG_YASDB_PROCESS_STATUS:  b.yasdbProcessStatus,
		data.DIAG_YASDB_INSTANCE_STATUS: b.yasdbInstanceStatus,
		data.DIAG_YASDB_DATABASE_STATUS: b.yasdbDatabaseStatus,
		data.DIAG_YASDB_ADR:             b.yasdbADRLog,
		data.DIAG_YASDB_ALERTLOG:        b.yasdbAlertLog,
		data.DIAG_YASDB_RUNLOG:          b.yasdbRunLog,
		data.DIAG_YASDB_COREDUMP:        b.yasdbCoredumpFile,
		data.DIAG_HOST_SYSTEMLOG:        b.hostSystemLog,
	}
}
