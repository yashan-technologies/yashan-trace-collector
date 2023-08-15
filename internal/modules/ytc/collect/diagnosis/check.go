package diagnosis

import (
	"os"
	"path"
	"strings"

	"ytc/defs/collecttypedef"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"
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
	corePattern := string(corePatternBytes)
	if !strings.HasPrefix(corePattern, "|") {
		if path.IsAbs(corePattern) {
			return path.Dir(corePattern), nil
		}
		return corePattern, err
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
	return path.Join(yasdbData, "log", "alert")
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
		m[data.DIAG_YASDB_RUNLOG] = path.Join(p, "run.log")
	}
	p = GetYasdbAlertLogPath(collectParam.YasdbData)
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_YASDB_ALERTLOG, err.Error())
	} else {
		m[data.DIAG_YASDB_ALERTLOG] = path.Join(p, "alert.log")
	}
	p, err = GetSystemLogPath()
	if err != nil {
		log.Module.Warnf(_getErrMessage, data.DIAG_HOST_SYSTEMLOG, err.Error())
	} else {
		m[data.DIAG_HOST_SYSTEMLOG] = p
	}
	return
}

func (b *DiagCollecter) checkYasdbEnv() error {
	env := yasdb.YasdbEnv{
		YasdbHome:     b.YasdbHome,
		YasdbData:     b.YasdbData,
		YasdbUser:     b.YasdbUser,
		YasdbPassword: b.YasdbPassword,
	}
	if err := env.ValidYasdbUserAndPwd(); err != nil {
		return err
	}
	return nil
}
