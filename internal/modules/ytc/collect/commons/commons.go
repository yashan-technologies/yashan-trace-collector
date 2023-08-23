package ytccollectcommons

import (
	"fmt"
	"os"
	"time"

	"ytc/defs/errdef"
	"ytc/log"
	"ytc/utils/userutil"
	"ytc/utils/yasqlutil"
)

const (
	fileNotExistDesc            = "%s not exist"
	fileNotPermissionDesc       = "current user: %s stat: %s permission denied"
	fileNotExistTips            = "you can check if %s exists"
	itemEmptyTips               = "you can re-run 'ytcctl collect' and fill yashandb user and yashandb password"
	fileNotPermissionTips       = "you can check whether the current user has access to %s,and use the user who belongs to this file to execute 'ytcctl collect'"
	databaseNotOpenTips         = "you can open the database first"
	invalidUserOrPasswordTips   = "you can enter a correct username and password"
	lackLoginPermissionTips     = "you can grant the user the 'Create Session' permission"
	lackNecessaryPermissionTips = "you can ask the database administrator or designated security administrator to grant the user the necessary privileges"
	objectNotExistTips          = "you can the object does not exist, please enter the correct object name"
	yasOtherTips                = "you can refer to the official documentation, view the error message, and then check the database"
	connectFailedTips           = "you can check process of yashandb"
	loginErr                    = "login yasdb failed: %s"
)

const (
	PLEASE_RUN_WITH_SUDO_TIPS = "you can run 'sudo ytcctl collect'"
	PLEASE_RUN_WITH_ROOT_TIPS = "you can run 'ytcctl collect' with root"
)

// base
const (
	YASDB_INSTANCE_STATUS_TIPS = "matched yasdb process with :%s ,default collect from process cmdline"
	GET_OSRELEASE_ERR_DESC     = "get os release err: %s"
	DEFAULT_PARAMETER_TIPS     = "default to collect parameter from %s"
)

// diag
const (
	DEFAULT_ADR_TIPS       = "default collect adr from: %s"
	MATCH_PROCESS_ERR_DESC = "match yasdb process with: %s err: %s"
	MATCH_PROCESS_ERR_TIPS = "you can try again later"
	PROCESS_NO_FOUND_DESC  = "process no found,match yasdb process with: %s"
	PROCESS_NO_FUNND_TIPS  = "you can check yasdb status"
	DEFAULT_RUNLOG_TIPS    = "default collect run.log from: %s"
	COREDUMP_ERR_DESC      = "get coredump path err: %s"
	COREDUMP_RELATIVE_DESC = "current core pattern: %s is relative path"
	COREDUMP_RELATIVE_TIPS = "default to: %s collect core file"
	GET_SYSLOG_ERR_DESC    = "get system err: %s"
	SYSLOG_UN_FOUND_DESC   = "both of %s and %s are not exist"
	SYSLOG_UN_FOUND_TIPS   = "do not collect system log"
	DMESG_NEED_ROOT_DESC   = "command dmesg need root"
)

// performance
const (
	USER_NOT_SYS_DESC = "the current yashandb user: %s is not 'SYS'"
	USER_NOT_SYS_TIPS = "you can change yashandb user to 'SYS'"

	DEFAULT_SLOWSQL_TIPS = "default collect slow.log from %s"
	AWR_TIMEOUT_DESC     = "it may take a long time to generate an AWR report."
	AWR_TIMEOUT_TIPS     = "we have defined the timeout period: %s for collecting AWR report, you can modify strategy.toml 'awr_timeout' to customize the timeout period"

	NO_SATISFIED_SNAP_DESC = "no satisfied snapshot id"
	NO_SATISFIED_TIPS      = "you can increase the collection interval appropriately"
)

// yasdb home
const (
	BIN   = "bin"
	YASQL = "yasql"
	YASDB = "yasdb"
)

// yasdb path
const (
	RUN    = "run"
	LOG    = "log"
	CONFIG = "config"
	SLOW   = "slow"
	DIAG   = "diag"
	ALERT  = "alert"
)

// yasdb node file
const (
	RUN_LOG   = "run.log"
	ALERT_LOG = "alert.log"
	SLOW_LOG  = "slow.log"
	YASDB_INI = "yasdb.ini"
)

type NoAccessRes struct {
	ModuleItem   string
	Description  string
	Tips         string
	ForceCollect bool // default false
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func PathErrDescAndTips(path string, e error) (desc, tips string) {
	if os.IsNotExist(e) {
		desc = fmt.Sprintf(fileNotExistDesc, path)
		tips = fmt.Sprintf(fileNotExistTips, path)
		return
	}
	if os.IsPermission(e) {
		desc = fmt.Sprintf(fileNotPermissionDesc, userutil.CurrentUser, path)
		if err := userutil.CheckSudovn(log.Module); err != nil {
			if err == userutil.ErrSudoNeedPwd {
				tips = PLEASE_RUN_WITH_SUDO_TIPS
				return
			}
			tips = PLEASE_RUN_WITH_ROOT_TIPS
			return
		}
		return
	}
	desc = e.Error()
	tips = " "
	return
}

func FillDescTips(no *NoAccessRes, desc, tips string) {
	if no == nil {
		return
	}
	no.Description = desc
	no.Tips = tips
}

func CheckSudoTips(err error) string {
	if err == nil {
		return ""
	}
	if err == userutil.ErrSudoNeedPwd {
		return PLEASE_RUN_WITH_SUDO_TIPS
	}
	return PLEASE_RUN_WITH_ROOT_TIPS
}

func YasErrDescAndtips(err error) (desc string, tips string) {
	if err == nil {
		return
	}
	desc = fmt.Sprintf(loginErr, err.Error())
	switch e := err.(type) {
	case *yasqlutil.YasErr:
		switch e.Prefix {
		case yasqlutil.YAS_DB_NOT_OPEN:
			tips = databaseNotOpenTips
		case yasqlutil.YAS_NO_DBUSER, yasqlutil.YAS_INVALID_USER_OR_PASSWORD:
			tips = invalidUserOrPasswordTips
		case yasqlutil.YAS_USER_LACK_LOGIN_AUTH:
			tips = lackLoginPermissionTips
		case yasqlutil.YAS_USER_LACK_AUTH:
			tips = lackNecessaryPermissionTips
		case yasqlutil.YAS_TABLE_OR_VIEW_DOES_NOT_EXIST:
			tips = objectNotExistTips
		case yasqlutil.YAS_FAILED_CONNECT_SOCKET:
			tips = connectFailedTips
		default:
			tips = yasOtherTips
		}
	case *errdef.ItemEmpty:
		tips = itemEmptyTips
	default:
		tips = " "
	}
	return
}

func NotAccessItem2Set(noAccess []NoAccessRes) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, noAccessRes := range noAccess {
		if noAccessRes.ForceCollect {
			continue
		}
		res[noAccessRes.ModuleItem] = struct{}{}
	}
	return
}
