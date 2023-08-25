package errdef

import (
	"errors"
	"fmt"
)

var (
	ErrPathFormat = errors.New("path format error, please check")
)

type ErrFileNotFound struct {
	Fname string
}

type ErrFileParseFailed struct {
	Fname string
	Err   error
}

type ErrCmdNotExist struct {
	Cmd string
}

type ErrCmdNeedRoot struct {
	Cmd string
}

type ErrPermissionDenied struct {
	User     string
	FileName string
}

func NewErrCmdNotExist(cmd string) *ErrCmdNotExist {
	return &ErrCmdNotExist{
		Cmd: cmd,
	}
}

func NewErrCmdNeedRoot(cmd string) *ErrCmdNeedRoot {
	return &ErrCmdNeedRoot{
		Cmd: cmd,
	}
}

func NewErrPermissionDenied(user string, path string) *ErrPermissionDenied {
	return &ErrPermissionDenied{
		User:     user,
		FileName: path,
	}
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("%s is not existed", e.Fname)
}

func (e *ErrFileParseFailed) Error() string {
	return fmt.Sprintf("parse %s failed: %s", e.Fname, e.Err)
}

func (e *ErrCmdNotExist) Error() string {
	return fmt.Sprintf("command: %s not exist", e.Cmd)
}

func (e *ErrCmdNeedRoot) Error() string {
	return fmt.Sprintf("command: %s need run with sudo or root", e.Cmd)
}

func (e *ErrPermissionDenied) Error() string {
	return fmt.Sprintf("The current user %s does not have permission to: %s", e.User, e.FileName)
}
