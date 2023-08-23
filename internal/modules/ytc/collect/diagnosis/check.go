package diagnosis

import (
	"fmt"
	"os"
	"path"
	"strings"

	"ytc/defs/collecttypedef"
	"ytc/defs/runtimedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
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
	return path.Join(yasdbData, ytccollectcommons.LOG, ytccollectcommons.ALERT)
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
		log.Module.Warnf(_getErrMessage, datadef.DIAG_YASDB_ADR, err.Error())
	} else {
		m[datadef.DIAG_YASDB_ADR] = p
	}
	p, err = GetCoredumpPath()
	if err != nil {
		log.Module.Warnf(_getErrMessage, datadef.DIAG_YASDB_COREDUMP, err.Error())
	} else {
		m[datadef.DIAG_YASDB_COREDUMP] = p
	}
	p, err = GetYasdbRunLogPath(collectParam)
	if err != nil {
		log.Module.Warnf(_getErrMessage, datadef.DIAG_YASDB_RUNLOG, err.Error())
	} else {
		m[datadef.DIAG_YASDB_RUNLOG] = path.Join(p, ytccollectcommons.RUN_LOG)
	}
	p = GetYasdbAlertLogPath(collectParam.YasdbData)
	if err != nil {
		log.Module.Warnf(_getErrMessage, datadef.DIAG_YASDB_ALERTLOG, err.Error())
	} else {
		m[datadef.DIAG_YASDB_ALERTLOG] = path.Join(p, ytccollectcommons.ALERT_LOG)
	}
	p, err = GetSystemLogPath()
	if err != nil {
		log.Module.Warnf(_getErrMessage, datadef.DIAG_HOST_SYSTEMLOG, err.Error())
	} else {
		m[datadef.DIAG_HOST_SYSTEMLOG] = p
	}
	return
}

func (d *DiagCollecter) checkYasdbProcess() *ytccollectcommons.NoAccessRes {
	proces, err := processutil.GetYasdbProcess(d.YasdbData)
	if err != nil || len(proces) == 0 {
		var (
			desc  string
			tips  string
			force bool
		)
		if err != nil {
			desc = fmt.Sprintf(ytccollectcommons.MATCH_PROCESS_ERR_DESC, d.YasdbData, err.Error())
			tips = ytccollectcommons.MATCH_PROCESS_ERR_TIPS
			force = true
		}
		if len(proces) == 0 {
			desc = fmt.Sprintf(ytccollectcommons.PROCESS_NO_FOUND_DESC, d.YasdbData)
			tips = ytccollectcommons.PROCESS_NO_FUNND_TIPS
		}
		return &ytccollectcommons.NoAccessRes{
			ModuleItem:   datadef.DIAG_YASDB_PROCESS_STATUS,
			Description:  desc,
			Tips:         tips,
			ForceCollect: force,
		}
	}
	return nil
}

func (d *DiagCollecter) checkYasdbInstanceStatus() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_YASDB_INSTANCE_STATUS
	yasql := path.Join(d.YasdbHome, ytccollectcommons.BIN, ytccollectcommons.YASQL)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(yasql, err)
		procs, processErr := processutil.GetYasdbProcess(d.YasdbData)
		if processErr != nil || len(procs) == 0 {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		tips = fmt.Sprintf(ytccollectcommons.YASDB_INSTANCE_STATUS_TIPS, d.YasdbData)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		desc, tips := ytccollectcommons.YasErrDescAndtips(d.yasdbValidateErr)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbDatabaseStatus() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_YASDB_DATABASE_STATUS
	yasql := path.Join(d.YasdbHome, ytccollectcommons.BIN, ytccollectcommons.YASQL)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(yasql, err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		desc, tips := ytccollectcommons.YasErrDescAndtips(d.yasdbValidateErr)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbAdr() *ytccollectcommons.NoAccessRes {
	diag := path.Join(d.YasdbData, ytccollectcommons.DIAG)
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_YASDB_ADR
	yasql := path.Join(d.YasdbHome, ytccollectcommons.BIN, ytccollectcommons.YASQL)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(yasql, err)
		if dErr := fileutil.CheckAccess(diag); dErr != nil {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		ytccollectcommons.FillDescTips(noAccess, desc, fmt.Sprintf(ytccollectcommons.DEFAULT_ADR_TIPS, diag))
		noAccess.ForceCollect = true
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		d.notConnectDB = true
		desc, tips := ytccollectcommons.YasErrDescAndtips(d.yasdbValidateErr)
		if dErr := fileutil.CheckAccess(diag); dErr != nil {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		noAccess.ForceCollect = true
		ytccollectcommons.FillDescTips(noAccess, desc, fmt.Sprintf(ytccollectcommons.DEFAULT_ADR_TIPS, diag))
		return noAccess
	}
	adrPath, err := GetAdrPath(d.CollectParam)
	if err != nil {
		d.notConnectDB = true
		desc, tips := ytccollectcommons.YasErrDescAndtips(err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	if err := fileutil.CheckAccess(adrPath); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(adrPath, err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbRunLog() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_YASDB_RUNLOG
	yasql := path.Join(d.YasdbHome, ytccollectcommons.BIN, ytccollectcommons.YASQL)
	defaultRunLog := path.Join(d.YasdbData, ytccollectcommons.LOG, ytccollectcommons.RUN, ytccollectcommons.RUN_LOG)
	err := fileutil.CheckAccess(yasql)
	if err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(yasql, err)
		if dErr := fileutil.CheckAccess(defaultRunLog); dErr != nil {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		ytccollectcommons.FillDescTips(noAccess, desc, fmt.Sprintf(ytccollectcommons.DEFAULT_RUNLOG_TIPS, defaultRunLog))
		noAccess.ForceCollect = true
		return noAccess
	}
	if d.yasdbValidateErr != nil {
		d.notConnectDB = true
		desc, tips := ytccollectcommons.YasErrDescAndtips(d.yasdbValidateErr)
		if dErr := fileutil.CheckAccess(defaultRunLog); dErr != nil {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		tips = fmt.Sprintf(ytccollectcommons.DEFAULT_RUNLOG_TIPS, defaultRunLog)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		noAccess.ForceCollect = true
		return noAccess
	}
	runLogPath, err := GetYasdbRunLogPath(d.CollectParam)
	if err != nil {
		d.notConnectDB = true
		desc, tips := ytccollectcommons.YasErrDescAndtips(err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	runLog := path.Join(runLogPath, ytccollectcommons.RUN_LOG)
	if err := fileutil.CheckAccess(runLog); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(runLog, err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbAlertLog() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_YASDB_ALERTLOG
	alertLogPath := GetYasdbAlertLogPath(d.YasdbData)
	alertLog := path.Join(alertLogPath, ytccollectcommons.ALERT_LOG)
	if err := fileutil.CheckAccess(alertLog); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(alertLog, err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkYasdbCoredump() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_YASDB_COREDUMP
	core, err := GetCoredumpPath()
	if err != nil {
		noAccess.Description = fmt.Sprintf(ytccollectcommons.COREDUMP_ERR_DESC, err.Error())
		noAccess.Tips = " "
		return noAccess
	}
	if !path.IsAbs(core) {
		bin := path.Join(d.YasdbHome, ytccollectcommons.BIN)
		if err := fileutil.CheckAccess(bin); err != nil {
			desc, tips := ytccollectcommons.PathErrDescAndTips(core, err)
			noAccess.Description = desc
			noAccess.Tips = tips
			return noAccess
		}
		noAccess.Description = fmt.Sprintf(ytccollectcommons.COREDUMP_RELATIVE_DESC, core)
		noAccess.Tips = fmt.Sprintf(ytccollectcommons.COREDUMP_RELATIVE_TIPS, bin)
		noAccess.ForceCollect = true
		return noAccess
	}
	if err := fileutil.CheckAccess(core); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(core, err)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkSyslog() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_HOST_SYSTEMLOG
	path, err := GetSystemLogPath()
	if err != nil {
		noAccess.Description = fmt.Sprintf(ytccollectcommons.GET_SYSLOG_ERR_DESC, err.Error())
		noAccess.Tips = " "
		return noAccess
	}
	if len(path) == 0 {
		noAccess.Description = fmt.Sprintf(ytccollectcommons.SYSLOG_UN_FOUND_DESC, SYSTEM_LOG_MESSAGES, SYSTEM_LOG_SYSLOG)
		noAccess.Tips = ytccollectcommons.SYSLOG_UN_FOUND_TIPS
		return noAccess
	}
	if err := fileutil.CheckAccess(path); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(path, err)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkDmesg() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_HOST_KERNELLOG
	release := runtimedef.GetOSRelease()
	if release.Id == osutil.KYLIN_ID {
		noAccess.Description = ytccollectcommons.DMESG_NEED_ROOT_DESC
		if sudoErr := userutil.CheckSudovn(log.Module); sudoErr != nil {
			noAccess.Tips = ytccollectcommons.PLEASE_RUN_WITH_ROOT_TIPS
			return noAccess
		}
		noAccess.Tips = ytccollectcommons.PLEASE_RUN_WITH_SUDO_TIPS
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) CheckFunc() map[string]checkFunc {
	return map[string]checkFunc{
		datadef.DIAG_YASDB_PROCESS_STATUS:  d.checkYasdbProcess,
		datadef.DIAG_YASDB_INSTANCE_STATUS: d.checkYasdbInstanceStatus,
		datadef.DIAG_YASDB_DATABASE_STATUS: d.checkYasdbDatabaseStatus,
		datadef.DIAG_YASDB_ADR:             d.checkYasdbAdr,
		datadef.DIAG_YASDB_RUNLOG:          d.checkYasdbRunLog,
		datadef.DIAG_YASDB_ALERTLOG:        d.checkYasdbAlertLog,
		datadef.DIAG_YASDB_COREDUMP:        d.checkYasdbCoredump,
		datadef.DIAG_HOST_SYSTEMLOG:        d.checkSyslog,
		datadef.DIAG_HOST_KERNELLOG:        d.checkDmesg,
	}
}
