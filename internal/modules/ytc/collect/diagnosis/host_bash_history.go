package diagnosis

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"ytc/defs/bashdef"
	"ytc/defs/runtimedef"
	"ytc/defs/timedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"
	"ytc/utils/userutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/execer"
)

const (
	hash                  = "#"
	bash_history_filename = ".bash_history"
	bash_history_format   = "%d\t%s\t%s\n"
	timestamp_placeholder = "<timestamp-missed>"
	bash_history_ctl      = "bashhistoryctl.sh"
)

const (
	has_no_permission_message     = "has no permission to collect bash history"
	has_no_permission_description = "没有权限收集Bash历史记录"
)

const (
	bhp_has_no_permission bashHistoryPermission = iota
	bhp_has_root_permission
	bhp_has_sudo_permission
	bhp_can_su_to_root_without_password_permission
	bhp_can_su_to_yasdb_user_without_password_permission
)

var (
	_currentBashHistoryPermission bashHistoryPermission

	_bashHistoryPermissionNames = map[bashHistoryPermission]string{
		bhp_has_no_permission:                                "has no permission to collect the bash history",
		bhp_has_root_permission:                              "has 'root' permission to collect the bash history",
		bhp_has_sudo_permission:                              "has 'sudo' permission to collect the bash history",
		bhp_can_su_to_root_without_password_permission:       "has 'su to root without password' permission to collect the bash history",
		bhp_can_su_to_yasdb_user_without_password_permission: "has 'su to yasdb user without password' permission to collect the bash history",
	}
)

type bashHistoryPermission int

type bashHistory struct {
	number  int
	time    string
	command string
}

func readBashHistoryFile(bashHistoryFile, timeFormat string) (content string, err error) {
	if _, err = os.Stat(bashHistoryFile); err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		return
	}

	f, err := os.Open(bashHistoryFile)
	if err != nil {
		return
	}
	defer f.Close()

	var historys []bashHistory
	history := bashHistory{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, hash) {
			timestamp, _ := strconv.ParseInt(strings.TrimPrefix(line, hash), 10, 64)
			if timestamp == 0 {
				continue
			}
			history.time = time.Unix(timestamp, 0).Format(timeFormat)
			continue
		}
		history.command = line
		history.number = len(historys) + 1
		if len(history.time) == 0 {
			history.time = timestamp_placeholder
		}
		historys = append(historys, history)
	}
	for _, h := range historys {
		content += fmt.Sprintf(bash_history_format, h.number, h.time, h.command)
	}
	return
}

func (d *DiagCollecter) genTargetUsers() map[string]struct{} {
	users := make(map[string]struct{})
	users[runtimedef.GetRootUsername()] = struct{}{}
	users[d.YasdbHomeOSUser] = struct{}{}
	return users
}

func (d *DiagCollecter) genBashHistoryFileName(destPath, user string) string {
	return fmt.Sprintf("%s/%s-bashhistory.txt", destPath, user)
}

func (d *DiagCollecter) genBashHistoryTmpFileName(user string) string {
	return fmt.Sprintf("/tmp/%s-%s.bashhistory", user, d.GetPackageTimestamp())
}

func (d *DiagCollecter) genDumpBashHistoryTmpFileCommand(script, user string) string {
	return fmt.Sprintf("%s dump %s", script, d.genBashHistoryTmpFileName(user))
}

func (d *DiagCollecter) collectHostBashHistoryByPermission(logger yaslog.YasLog, users map[string]struct{}, destPath, script string) map[string]string {
	resp := make(map[string]string)
	executor := execer.NewExecer(logger)
	for user := range users {
		var bin string
		var args []string
		if user == userutil.CurrentUser {
			bin = bashdef.CMD_BASH
			args = []string{"-c", d.genDumpBashHistoryTmpFileCommand(script, user)}
		} else {
			switch _currentBashHistoryPermission {
			case bhp_has_root_permission, bhp_can_su_to_yasdb_user_without_password_permission:
				bin = bashdef.CMD_SU
				args = []string{"-l", user, "-c", d.genDumpBashHistoryTmpFileCommand(script, user)}
			case bhp_has_sudo_permission:
				bin = bashdef.CMD_SUDO
				args = []string{bashdef.CMD_SU, "-l", user, "-c", d.genDumpBashHistoryTmpFileCommand(script, user)}
			case bhp_can_su_to_root_without_password_permission:
				bin = bashdef.CMD_SU
				args = []string{"-l", runtimedef.GetRootUsername(), "-c", bashdef.CMD_SU, user, "-c", d.genDumpBashHistoryTmpFileCommand(script, user)}
			}
		}
		if _currentBashHistoryPermission == bhp_can_su_to_yasdb_user_without_password_permission &&
			user != d.YasdbHomeOSUser {
			continue
		}
		ret, _, stderr := executor.Exec(bin, args...)
		if ret != 0 {
			resp[user] = stderr
			continue
		}
		content, err := readBashHistoryFile(d.genBashHistoryTmpFileName(user), timedef.TIME_FORMAT)
		if err != nil {
			resp[user] = err.Error()
			continue
		}
		if err := fileutil.WriteFile(d.genBashHistoryFileName(destPath, user), []byte(content)); err != nil {
			resp[user] = err.Error()
			continue
		}
		_ = os.Remove(d.genBashHistoryTmpFileName(user))
	}
	return resp
}

func (d *DiagCollecter) collectHostBashHistory() (err error) {
	destPath := path.Join(_packageDir, ytccollectcommons.HOST_DIR_NAME, BASH_HISTORY_DIR_NAME)
	script := path.Join(runtimedef.GetScriptsPath(), bash_history_ctl)
	resp := datadef.YTCItem{
		Name:    datadef.DIAG_HOST_BASH_HISTORY,
		Details: d.GenPackageRelativePath(path.Join(ytccollectcommons.HOST_DIR_NAME, BASH_HISTORY_DIR_NAME)),
	}
	defer d.fillResult(&resp)
	logger := log.Module.M(datadef.DIAG_HOST_BASH_HISTORY)

	users := d.genTargetUsers()
	logger.Infof("bash history permission: %s", _bashHistoryPermissionNames[_currentBashHistoryPermission])
	switch _currentBashHistoryPermission {
	case bhp_has_root_permission, bhp_has_sudo_permission,
		bhp_can_su_to_root_without_password_permission, bhp_can_su_to_yasdb_user_without_password_permission:
		errorMap := d.collectHostBashHistoryByPermission(logger, users, destPath, script)
		var errs []string
		for user, err := range errorMap {
			message := fmt.Sprintf("collect %s bash history failed: %s", user, err)
			errs = append(errs, err)
			logger.Error(message)
		}
		if len(errorMap) == len(users) {
			resp.Error = strings.Join(errs, stringutil.STR_COMMA)
			resp.Description = datadef.GenDefaultDesc()
		}
	case bhp_has_no_permission:
		resp.Error = has_no_permission_message
		resp.Description = has_no_permission_description
		return
	}
	return
}
