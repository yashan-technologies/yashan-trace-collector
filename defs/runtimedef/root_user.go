package runtimedef

import (
	"ytc/utils/userutil"
)

const DEFAULT_ROOT_USERNAME = "root"

var _rootUserName string

func initRootUsername() {
	u, err := userutil.GetRootUser()
	if err != nil {
		_rootUserName = DEFAULT_ROOT_USERNAME
		return
	}
	_rootUserName = u.Username
}

func GetRootUsername() string {
	return _rootUserName
}
