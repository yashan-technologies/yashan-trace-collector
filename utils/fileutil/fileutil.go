// The stringutil package encapsulates functions related to files.
package fileutil

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

var (
	ErrSyscallStatNotSupported = errors.New("syscall stat not supported")
)

type Owner struct {
	Uid       int
	Gid       int
	Username  string
	GroupName string
}

// IsPathSymlink checks whether a path is a real path or a link, and returns the real path when it is a link.
func IsPathSymlink(path string) (isSymlink bool, realPath string, err error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return
	}
	isSymlink = fi.Mode()&os.ModeSymlink != 0
	if isSymlink {
		realPath, err = os.Readlink(path)
	}
	return
}

// GetRealPath returns the real path directly when path is a real path,
// and returns the real path pointed to by a link when path is a link.
func GetRealPath(path string) (realPath string, err error) {
	return filepath.EvalSymlinks(path)
}

// GetOwnerID gets the user ID and user group ID to which path belongs.
func GetOwnerID(path string) (uid uint32, gid uint32, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	state, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		err = ErrSyscallStatNotSupported
		return
	}
	uid, gid = state.Uid, state.Gid
	return
}

// GetOwnerID gets the username and user group name to which path belongs.
func GetOwner(path string) (owner Owner, err error) {
	uid, gid, err := GetOwnerID(path)
	if err != nil {
		return
	}
	u, err := user.LookupId(fmt.Sprint(uid))
	if err != nil {
		return
	}
	g, err := user.LookupGroupId(fmt.Sprint(gid))
	if err != nil {
		return
	}
	owner = Owner{
		Uid:       int(uid),
		Gid:       int(gid),
		Username:  u.Username,
		GroupName: g.Name,
	}
	return
}
