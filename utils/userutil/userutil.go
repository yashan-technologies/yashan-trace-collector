// The userutil package encapsulates functions related to users and user groups.
package userutil

import (
	"os"
	"os/user"
	"strconv"
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
