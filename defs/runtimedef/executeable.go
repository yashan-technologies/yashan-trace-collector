package runtimedef

import (
	"os"

	"ytc/utils/fileutil"
)

var (
	_owner fileutil.Owner
)

func GetExecuteableOwner() fileutil.Owner {
	return _owner
}

func getExecutable() (executeable string, err error) {
	executeable, err = os.Executable()
	if err != nil {
		return
	}
	return fileutil.GetRealPath(executeable)
}

func setOwner(owner fileutil.Owner) {
	_owner = fileutil.Owner{
		Uid:       owner.Uid,
		Gid:       owner.Gid,
		Username:  owner.Username,
		GroupName: owner.GroupName,
	}
}

func initExecuteable() (err error) {
	executeable, err := getExecutable()
	if err != nil {
		return
	}
	owner, err := fileutil.GetOwner(executeable)
	if err != nil {
		return
	}
	setOwner(owner)
	return
}
