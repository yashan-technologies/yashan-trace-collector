package ytccollectcommons

import (
	"fmt"
	"os"
	"ytc/defs/errdef"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/log"
	"ytc/utils/userutil"
	"ytc/utils/yasqlutil"
)

type FilePathErrType string

const (
	NotExist      FilePathErrType = "NotExist"
	NotPermission FilePathErrType = "NotPermission"
)

const (
	fileNotExistDesc            = "%s not exist"
	fileNotPermissionDesc       = "current user: %s stat: %s permission denied"
	fileNotExistTips            = "you can check if %s exists"
	itemEmptyTips               = "you can re-run 'ytcctl collect' and full yashandb user and yashandb password"
	fileNotPermissionTips       = "you can check whether the current user has access to %s,and use the user who belongs to this file to execute 'ytcctl collect'"
	databaseNotOpenTips         = "you can open the database first"
	invalidUserOrPasswordTips   = "you can enter a correct username and password"
	lackLoginPermissionTips     = "you can grant the user the 'Create Session' permission"
	lackNecessaryPermissionTips = "you can ask the database administrator or designated security administrator to grant the user the necessary privileges"
	objectNotExistTips          = "you can the object does not exist, please enter the correct object name"
	yasOtherTips                = "you can refer to the official documentation, view the error message, and then check the database"
	connectFailedTips           = "you can check process of yashandb"
	loginErr                    = "login yasdb failed: %s"

	PLEASE_RUN_WITH_SUDO_TIPS = "you can run 'sudo ytcctl collect'"
	PLEASE_RUN_WITH_ROOT_TIPS = "you can run 'ytcctl collect' with root"
)

// base
const (
	YasdbInstanceStatusTips = "matched yasdb process with :%s ,default collect from process cmdline"
	GetOsReleaseErrDesc     = "get os release err: %s"
	DefaultParameterTips    = "default to collect parameter from %s"
)

// diag
const (
	DefaultAdrTips       = "default collect adr from: %s"
	MatchProcessErrDesc  = "match yasdb process with: %s err: %s"
	MatchProcessErrTips  = "you can try again later"
	ProcessNofoundDesc   = "process no found,match yasdb process with: %s"
	ProcessNofunndTips   = "you can check yasdb status"
	DefaultRunlogTips    = "default collect run.log from: %s"
	CoredumpErrDesc      = "get coredump path err: %s"
	CoredumpRelativeDesc = "current core pattern: %s is relative path"
	CoredumpRelativeTips = "default to: %s collect core file"
	GetSysLogErrDesc     = "get system err: %s"
	SysLogUnfoundDesc    = "both of %s and %s are not exist"
	SysLogUnfoundTips    = "do not collect system log"
	DmesgNeedRootDesc    = "command dmesg need root"
)

const (
	BIN       = "bin"
	YASQL     = "yasql"
	YASDB     = "yasdb"
	LOG       = "log"
	RUN       = "run"
	RUN_LOG   = "run.log"
	ALERT     = "alert"
	ALERT_LOG = "alert.log"
	CONFIG    = "config"
	DIAG      = "diag"
	YASDB_INI = "yasdb.ini"
)

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

func FillDescTips(no *data.NoAccessRes, desc, tips string) {
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
