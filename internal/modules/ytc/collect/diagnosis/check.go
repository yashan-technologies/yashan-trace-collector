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
	return strings.ReplaceAll(dest, stringutil.STR_QUESTION_MARK, collectParam.YasdbData), err
}

func GetCoreDumpPath() (string, string, error) {
	corePatternBytes, err := fileutil.ReadFile(CORE_PATTERN_PATH)
	if err != nil {
		return "", "", err
	}
	corePattern := strings.TrimSpace(string(corePatternBytes))
	if !strings.HasPrefix(corePattern, stringutil.STR_BAR) {
		return corePattern, CORE_DIRECT, nil
	}
	if strings.Contains(corePattern, ABRT_HOOK_CPP) {
		localtion, err := fileutil.GetConfByKey(ABRT_CONF, KEY_DUMP_LOCATION)
		if err != nil {
			log.Module.Errorf("get %s from %s err:%s", KEY_DUMP_LOCATION, ABRT_CONF, err.Error())
			return "", "", err
		}
		if stringutil.IsEmpty(localtion) {
			localtion = DEFAULT_DUMP_LOCATION
		}
		return localtion, CORE_REDIRECT_ABRT, nil
	}
	if strings.Contains(corePattern, SYSTEMD_COREDUMP) {
		storage, err := fileutil.GetConfByKey(SYSTEMD_COREDUMP_CONF, KEY_STORAGE)
		if err != nil {
			log.Module.Errorf("get %s from %s err:%s", SYSTEMD_COREDUMP_CONF, KEY_STORAGE, err.Error())
			return "", "", err
		}
		// do not collect coredump
		if storage != VALUE_EXTERNAL {
			err := fmt.Errorf("the host coredump config is %s, does not collect", storage)
			log.Module.Error(err)
			return "", "", err
		}
		return DEFAULT_EXTERNAL_STORAGE, CORE_REDIRECT_SYSTEMD, nil
	}
	err = fmt.Errorf("core parttern %s is unknown, do not collect", corePattern)
	log.Module.Error(err)
	return "", "", err
}

func GetYasdbRunLogPath(collectParam *collecttypedef.CollectParam) (string, error) {
	tx := yasqlutil.GetLocalInstance(collectParam.YasdbUser, collectParam.YasdbPassword, collectParam.YasdbHome, collectParam.YasdbData)
	dest, err := yasdb.QueryParameter(tx, yasdb.PM_RUN_LOG_FILE_PATH)
	return strings.ReplaceAll(dest, stringutil.STR_QUESTION_MARK, collectParam.YasdbData), err
}

func GetYasdbAlertLogPath(yasdbData string) string {
	return path.Join(yasdbData, ytccollectcommons.LOG, ytccollectcommons.ALERT)
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
	res, err := fileutil.CheckDirAccess(adrPath, nil)
	if err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(adrPath, err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	if len(res) != 0 {
		desc, tips := ytccollectcommons.FilesErrDescAndTips(res)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		noAccess.ForceCollect = true
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
	originCoreDumpPath, coreDumpType, err := GetCoreDumpPath()
	if err != nil {
		noAccess.Description = fmt.Sprintf(ytccollectcommons.COREDUMP_ERR_DESC, err.Error())
		noAccess.Tips = " "
		return noAccess
	}
	coreDumpPath := d.getCoreDumpRealPath(originCoreDumpPath, coreDumpType)
	if err := fileutil.CheckAccess(coreDumpPath); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(coreDumpPath, err)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	// gen relative path info
	if !path.IsAbs(originCoreDumpPath) {
		noAccess.Description = fmt.Sprintf(ytccollectcommons.COREDUMP_RELATIVE_DESC, originCoreDumpPath)
		noAccess.Tips = fmt.Sprintf(ytccollectcommons.COREDUMP_RELATIVE_TIPS, coreDumpPath)
		noAccess.ForceCollect = true
		return noAccess
	}
	return nil
}

func (d *DiagCollecter) checkSyslog() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_HOST_SYSTEMLOG
	messageErr := fileutil.CheckAccess(SYSTEM_LOG_MESSAGES)
	syslogErr := fileutil.CheckAccess(SYSTEM_LOG_SYSLOG)
	if messageErr == nil {
		return nil
	}
	if !os.IsNotExist(messageErr) {
		desc, tips := ytccollectcommons.PathErrDescAndTips(SYSTEM_LOG_MESSAGES, messageErr)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	if syslogErr == nil {
		return nil
	}
	if !os.IsNotExist(syslogErr) {
		desc, tips := ytccollectcommons.PathErrDescAndTips(SYSTEM_LOG_SYSLOG, syslogErr)
		noAccess.Description = desc
		noAccess.Tips = tips
		return noAccess
	}
	noAccess.Description = fmt.Sprintf(ytccollectcommons.SYSLOG_UN_FOUND_DESC, SYSTEM_LOG_MESSAGES, SYSTEM_LOG_SYSLOG)
	noAccess.Tips = ytccollectcommons.SYSLOG_UN_FOUND_TIPS
	return noAccess
}

func (d *DiagCollecter) checkDmesg() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_HOST_KERNELLOG
	release := runtimedef.GetOSRelease()
	if release.Id == osutil.KYLIN_ID {
		if !userutil.IsCurrentUserRoot() {
			noAccess.Description = ytccollectcommons.DMESG_NEED_ROOT_DESC
			if sudoErr := userutil.CheckSudovn(log.Module); sudoErr != nil {
				noAccess.Tips = ytccollectcommons.PLEASE_RUN_WITH_ROOT_TIPS
				return noAccess
			}
			noAccess.Tips = ytccollectcommons.PLEASE_RUN_WITH_SUDO_TIPS
			return noAccess
		}
	}
	return nil
}

func (d *DiagCollecter) checkBashHistory() *ytccollectcommons.NoAccessRes {
	const default_description = "The data source of bash history is the $HISTFILE file, " +
		"if you need to keep the history of the current terminal, please execute 'history -a' first."
	logger := log.Module.M(datadef.DIAG_HOST_BASH_HISTORY)
	resp := &ytccollectcommons.NoAccessRes{
		ModuleItem:   datadef.DIAG_HOST_BASH_HISTORY,
		Description:  default_description,
		ForceCollect: true,
	}

	if userutil.IsCurrentUserRoot() {
		_currentBashHistoryPermission = bhp_has_root_permission
		return resp
	}

	switch err := userutil.CheckSudovn(logger); err {
	case nil:
		_currentBashHistoryPermission = bhp_has_sudo_permission
		return resp
	case userutil.ErrSudoNeedPwd:
		resp.Description = err.Error()
		resp.Tips = ytccollectcommons.PLEASE_RUN_WITH_SUDO_TIPS
	default:
		resp.Description = err.Error()
		resp.Tips = ytccollectcommons.PLEASE_RUN_WITH_ROOT_TIPS
	}

	if userutil.CanSuToTargetUserWithoutPassword(runtimedef.GetRootUsername(), logger) {
		_currentBashHistoryPermission = bhp_can_su_to_root_without_password_permission
		resp.Description = default_description
		resp.Tips = ""
		return resp
	}
	resp.Description = "has no permission to collect bash history of root"

	canSuToYasdbUserWithoutPassword := false
	if d.CollectParam.YasdbHomeOSUser == userutil.CurrentUser {
		canSuToYasdbUserWithoutPassword = true
	} else {
		canSuToYasdbUserWithoutPassword = userutil.CanSuToTargetUserWithoutPassword(d.CollectParam.YasdbHomeOSUser, logger)
	}
	if canSuToYasdbUserWithoutPassword {
		_currentBashHistoryPermission = bhp_can_su_to_yasdb_user_without_password_permission
		return resp
	}
	resp.Description += fmt.Sprintf(" and %s", d.CollectParam.YasdbHomeOSUser)
	resp.ForceCollect = false
	return resp
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
		datadef.DIAG_HOST_BASH_HISTORY:     d.checkBashHistory,
	}
}
