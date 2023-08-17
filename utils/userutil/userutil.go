// The userutil package encapsulates functions related to users and user groups.
package userutil

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"ytc/defs/bashdef"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/execer"
)

const (
	NotRunSudo         = "not run sudo"
	PasswordIsRequired = "password is required"
)

var (
	ErrSudoNeedPwd = errors.New("a password is required")
	ErrNotRunSudo  = errors.New("user may not run sudo")
)

// GetUsernameById returns username by user ID.
func GetUsernameById(id int) (username string, err error) {
	u, err := user.LookupId(strconv.FormatInt(int64(id), 10))
	if err != nil {
		return
	}
	username = u.Username
	return
}

// GetCurrentUser returns the current username.
func GetCurrentUser() (string, error) {
	return GetUsernameById(os.Getuid())
}

// IsCurrentUserRoot checks whether the current user is root.
func IsCurrentUserRoot() bool {
	return os.Getuid() == 0
}

// IsSysUserExists checks if the OS user exists.
func IsSysUserExists(username string) bool {
	_, err := user.Lookup(username)
	return err == nil
}

// IsSysGroupExists checks if the OS user group exists.
func IsSysGroupExists(groupname string) bool {
	_, err := user.LookupGroup(groupname)
	return err == nil
}

// 通过sudo -vn命令判断sudo权限是否具备，预期会报错需要密码，则表示该用户具备sudo权限。
func CheckSudovn(logger yaslog.YasLog) error {
	exec := execer.NewExecer(logger)
	ret, _, err := exec.Exec(bashdef.CMD_BASH, "-c", fmt.Sprintf("%s %s", bashdef.CMD_SUDO, "-vn"))
	if ret == 0 {
		return nil
	}
	if strings.Contains(err, PasswordIsRequired) {
		return ErrSudoNeedPwd
	}
	if strings.Contains(err, NotRunSudo) {
		return ErrNotRunSudo
	}
	return errors.New(err)
}
