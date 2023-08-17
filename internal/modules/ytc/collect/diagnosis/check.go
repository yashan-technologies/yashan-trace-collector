package diagnosis

import (
	"fmt"
	"os"
	"path"
	"strings"

	"ytc/defs/collecttypedef"
	ytccommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/osutil"
	"ytc/utils/processutil"
	"ytc/utils/stringutil"
	"ytc/utils/userutil"
	"ytc/utils/yasqlutil"
)

func GetAdrPath(collectParam *collecttypedef.CollectParam) (string, error) {
	tx := yasqlutil.GetLocalInstance(collectParam.YasdbUser, collectParam.YasdbPassword, collectParam.YasdbHome, collectParam.YasdbData)
	dest, err := yasdb.QueryParameter(tx, yasdb.PM_DIAGNOSTIC_DEST)
	return strings.ReplaceAll(dest, "?", collectParam.YasdbData), err
}

func GetCoredumpPath() (string, error) {
	corePatternBytes, err := fileutil.ReadFile(CORE_PATTERN_PATH)
	if err != nil {
		return "", err
	}
	corePattern := strings.TrimSpace(string(corePatternBytes))
	if !strings.HasPrefix(corePattern, "|") {
		if path.IsAbs(corePattern) {
			return path.Dir(corePattern), nil
		}
		return corePattern, nil
	}
	if strings.Contains(corePattern, ABRT_HOOK_CPP) {
		localtion, err := fileutil.GetConfByKey(ABRT_CONF, DUMP_LOCATION)
		if err != nil {
			log.Module.Errorf("get %s from %s err:%s", DUMP_LOCATION, ABRT_CONF, err.Error())
			return "", err
		}
		if stringutil.IsEmpty(localtion) {
			localtion = DEFAULT_DUMP_LOCATION
		}
		return localtion, nil
	}
	if strings.Contains(corePattern, SYSTEMD_COREDUMP) {
		storage, err := fileutil.GetConfByKey(SYSTEMD_COREDUMP_CONF, STORAGE)
		if err != nil {
			log.Module.Errorf("get %s from %s err:%s", SYSTEMD_COREDUMP_CONF, STORAGE, err.Error())
			return "", err
		}
		// do not collect core dump
		if storage != EXTERNAL_STORAGE {
			log.Module.Warnf("the host coredump config is closed")
		}
		externalStorage, err := fileutil.GetConfByKey(SYSTEMD_COREDUMP_CONF, EXTERNAL_STORAGE)
		if err != nil {
			log.Module.Errorf("get %s from %s err:%s", SYSTEMD_COREDUMP_CONF, EXTERNAL_STORAGE, err.Error())
			return "", err
		}
		if stringutil.IsEmpty(externalStorage) {
			externalStorage = DEFAULT_EXTERNAL_STORAGE
		}
		return externalStorage, nil
	}
	log.Module.Warnf("core parttern %s is un known, do not collect", corePattern)
	return "", nil
}

func GetYasdbRunLogPath(collectParam *collecttypedef.CollectParam) (string, error) {
	tx := yasqlutil.GetLocalInstance(collectParam.YasdbUser, collectParam.YasdbPassword, collectParam.YasdbHome, collectParam.YasdbData)
	dest, err := yasdb.QueryParameter(tx, yasdb.PM_RUN_LOG_FILE_PATH)
	return strings.ReplaceAll(dest, "?", collectParam.YasdbData), err
}

func GetYasdbAlertLogPath(yasdbData string) string {
	return path.Join(yasdbData, ytccommons.LOG, ytccommons.ALERT)
}

func GetSystemLogPath() (string, error) {
	_, err := os.Stat(SYSTEM_LOG_MESSAGES)
	if err == nil {
		return SYSTEM_LOG_MESSAGES, nil
	}
	log.Module.Errorf(err.Error())
	if os.IsPermission(err) {
		return SYSTEM_LOG_MESSAGES, nil
	}
	_, err = os.Stat(SYSTEM_LOG_SYSLOG)
	if err == nil {
		return SYSTEM_LOG_MESSAGES, nil
	}
	if err != nil {
		log.Module.Errorf(err.Error())
		if os.IsPermission(err) {
			return SYSTEM_LOG_SYSLOG, nil
		}
	}
	log.Module.Warnf("%s and %s not exist do not collect", SYSTEM_LOG_MESSAGES, SYSTEM_LOG_SYSLOG)
	return "", err
}

// return diag item path map key:diagitem value path
func GetDiagPath(collectParam *collecttypedef.CollectParam) (m map[string]string) {
	m = make(map[string]string)
	p, err := GetAdrPath(collectParam)
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_YASDB_ADR, err.Error())
	} else {
		m[data.DIAG_YASDB_ADR] = p
	}
	p, err = GetCoredumpPath()
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_YASDB_COREDUMP, err.Error())
	} else {
		m[data.DIAG_YASDB_COREDUMP] = p
	}
	p, err = GetYasdbRunLogPath(collectParam)
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_YASDB_RUNLOG, err.Error())
	} else {
		m[data.DIAG_YASDB_RUNLOG] = path.Join(p, ytccommons.RUN_LOG)
	}
	p = GetYasdbAlertLogPath(collectParam.YasdbData)
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_YASDB_ALERTLOG, err.Error())
	} else {
		m[data.DIAG_YASDB_ALERTLOG] = path.Join(p, ytccommons.ALERT_LOG)
	}
	p, err = GetSystemLogPath()
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_HOST_SYSTEMLOG, err.Error())
	} else {
		m[data.DIAG_HOST_SYSTEMLOG] = p
	}
	return
}

func (d *DiagCollecter) checkYasdbProcess() *data.NoAccessRes {
	proces, err := processutil.GetYasdbProcess(d.YasdbData)
	if err != nil || len(proces) == 0 {
		var (
			desc  string
			tips  string
			force bool
		)
		if err != nil {
			desc = fmt.Sprintf(ytccommons.MatchProcessErrDesc, d.YasdbData, err.Error())
			tips = ytccommons.MatchProcessErrTips
			force = true
		}
		if len(proces) == 0 {
			desc = fmt.Sprintf(ytccommons.ProcessNofoundDesc, d.YasdbData)
			tips = ytccommons.ProcessNofunndTips
		}
		return &data.NoAccessRes{
			ModuleItem:   data.DIAG_YASDB_PROCESS_STATUS,
			Description:  desc,
			Tips:         tips,
			ForceCollect: force,
		}
	}
	return nil
}

func (d *DiagCollecter) checkYasdbInstanceStatus() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_YASDB_INSTANCE_STATUS
	yasql := path.Join(d.YasdbHome, ytccommons.BIN, ytccommons.YASQL)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(yasql, err)
		procs, processErr := processutil.GetYasdbProcess(d.YasdbData)
		if processErr != nil || len(procs) == 0 {
			ytccommons.FullDescTips(noAccess, desc, tips)
			return noAccess
		}
		tips = fmt.Sprintf(ytccommons.YasdbInstanceStatusTips, d.YasdbData)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		desc, tips := ytccommons.YasErrDescAndtips(d.yasdbValidateErr)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbDatabaseStatus() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_YASDB_DATABASE_STATUS
	yasql := path.Join(d.YasdbHome, ytccommons.BIN, ytccommons.YASQL)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(yasql, err)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		desc, tips := ytccommons.YasErrDescAndtips(d.yasdbValidateErr)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbAdr() *data.NoAccessRes {
	diag := path.Join(d.YasdbData, ytccommons.DIAG)
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_YASDB_ADR
	yasql := path.Join(d.YasdbHome, ytccommons.BIN, ytccommons.YASQL)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(yasql, err)
		if dErr := fileutil.CheckAccess(diag); dErr != nil {
			ytccommons.FullDescTips(noAccess, desc, tips)
			return noAccess
		}
		ytccommons.FullDescTips(noAccess, desc, fmt.Sprintf(ytccommons.DefaultAdrTips, diag))
		noAccess.ForceCollect = true
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		d.notConnectDB = true
		desc, tips := ytccommons.YasErrDescAndtips(d.yasdbValidateErr)
		if dErr := fileutil.CheckAccess(diag); dErr != nil {
			ytccommons.FullDescTips(noAccess, desc, tips)
			return noAccess
		}
		noAccess.ForceCollect = true
		ytccommons.FullDescTips(noAccess, desc, fmt.Sprintf(ytccommons.DefaultAdrTips, diag))
		return noAccess
	}
	adrPath, err := GetAdrPath(d.CollectParam)
	if err != nil {
		d.notConnectDB = true
		desc, tips := ytccommons.YasErrDescAndtips(err)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	if err := fileutil.CheckAccess(adrPath); err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(adrPath, err)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbRunLog() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_YASDB_RUNLOG
	yasql := path.Join(d.YasdbHome, ytccommons.BIN, ytccommons.YASQL)
	defaultRunLog := path.Join(d.YasdbData, ytccommons.LOG, ytccommons.RUN, ytccommons.RUN_LOG)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(yasql, err)
		if dErr := fileutil.CheckAccess(defaultRunLog); dErr != nil {
			ytccommons.FullDescTips(noAccess, desc, tips)
			return noAccess
		}
		ytccommons.FullDescTips(noAccess, desc, fmt.Sprintf(ytccommons.DefaultRunlogTips, defaultRunLog))
		noAccess.ForceCollect = true
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		d.notConnectDB = true
		desc, tips := ytccommons.YasErrDescAndtips(d.yasdbValidateErr)
		if dErr := fileutil.CheckAccess(defaultRunLog); dErr != nil {
			ytccommons.FullDescTips(noAccess, desc, tips)
			return noAccess
		}
		tips = fmt.Sprintf(ytccommons.DefaultRunlogTips, defaultRunLog)
		ytccommons.FullDescTips(noAccess, desc, tips)
		noAccess.ForceCollect = true
		return noAccess
	}
	runLogPath, err := GetYasdbRunLogPath(d.CollectParam)
	if err != nil {
		d.notConnectDB = true
		desc, tips := ytccommons.YasErrDescAndtips(err)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	runLog := path.Join(runLogPath, ytccommons.RUN_LOG)
	if err := fileutil.CheckAccess(runLog); err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(runLog, err)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbAlertLog() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_YASDB_ALERTLOG
	alertLogPath := GetYasdbAlertLogPath(d.YasdbData)
	alertLog := path.Join(alertLogPath, ytccommons.ALERT_LOG)
	if err := fileutil.CheckAccess(alertLog); err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(alertLog, err)
		ytccommons.FullDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbCoredump() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_YASDB_COREDUMP
	core, err := GetCoredumpPath()
	if err != nil {
		noAccess.Description = fmt.Sprintf(ytccommons.CoredumpErrDesc, err.Error())
		noAccess.Tips = " "
		return noAccess
	}
	if !path.IsAbs(core) {
		bin := path.Join(d.YasdbHome, ytccommons.BIN)
		if err := fileutil.CheckAccess(bin); err != nil {
			desc, tips := ytccommons.PathErrDescAndTips(core, err)
			noAccess.Description = desc
			noAccess.Tips = tips
			return noAccess
		}
		noAccess.Description = fmt.Sprintf(ytccommons.CoredumpRelativeDesc, core)
		noAccess.Tips = fmt.Sprintf(ytccommons.CoredumpRelativeTips, bin)
		noAccess.ForceCollect = true
		return noAccess
	}
	if err := fileutil.CheckAccess(core); err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(core, err)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkSyslog() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_HOST_SYSTEMLOG
	path, err := GetSystemLogPath()
	if err != nil {
		noAccess.Description = fmt.Sprintf(ytccommons.GetSysLogErrDesc, err.Error())
		noAccess.Tips = " "
		return noAccess
	}
	if len(path) == 0 {
		noAccess.Description = fmt.Sprintf(ytccommons.SysLogUnfoundDesc, SYSTEM_LOG_MESSAGES, SYSTEM_LOG_SYSLOG)
		noAccess.Tips = ytccommons.SysLogUnfoundTips
		return noAccess
	}
	if err := fileutil.CheckAccess(path); err != nil {
		desc, tips := ytccommons.PathErrDescAndTips(path, err)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkDmesg() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_HOST_DMESG
	release, err := osutil.GetOsRelease()
	if err != nil {
		noAccess.Description = fmt.Sprintf(ytccommons.GetOsReleaseErrDesc, err.Error())
		noAccess.Tips = " "
		noAccess.ForceCollect = true
		return noAccess
	}
	if release.Id == osutil.KYLIN_ID {
		noAccess.Description = ytccommons.DmesgNeedRootDesc
		if sudoErr := userutil.CheckSudovn(log.Module); sudoErr != nil {
			noAccess.Tips = ytccommons.PLEASE_RUN_WITH_ROOT_TIPS
			return noAccess
		}
		noAccess.Tips = ytccommons.PLEASE_RUN_WITH_SUDO_TIPS
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) CheckFunc() map[string]checkFunc {
	return map[string]checkFunc{
		data.DIAG_YASDB_PROCESS_STATUS:  d.checkYasdbProcess,
		data.DIAG_YASDB_INSTANCE_STATUS: d.checkYasdbInstanceStatus,
		data.DIAG_YASDB_DATABASE_STATUS: d.checkYasdbDatabaseStatus,
		data.DIAG_YASDB_ADR:             d.checkYasdbAdr,
		data.DIAG_YASDB_RUNLOG:          d.checkYasdbRunLog,
		data.DIAG_YASDB_ALERTLOG:        d.checkYasdbAlertLog,
		data.DIAG_YASDB_COREDUMP:        d.checkYasdbCoredump,
		data.DIAG_HOST_SYSTEMLOG:        d.checkSyslog,
		data.DIAG_HOST_DMESG:            d.checkDmesg,
	}
}
