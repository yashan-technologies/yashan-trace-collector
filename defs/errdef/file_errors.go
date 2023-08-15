package errdef

import "fmt"

type ErrFileNotFound struct {
	Fname string
}

type ErrFileParseFailed struct {
	Fname string
	Err   error
}

type ErrFlagFormat struct {
	Cmd  string
	Flag string
}

type ErrCmdNotExist struct {
	Cmd string
}

type ErrCmdNeedRoot struct {
	Cmd string
}

type ErrPermissionDenied struct {
	FileName string
}

func NewErrFlagFormat(cmd, flag string) *ErrFlagFormat {
	return &ErrFlagFormat{
		Cmd:  cmd,
		Flag: flag,
	}
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

func NewErrPermissionDenied(name string) *ErrPermissionDenied {
	return &ErrPermissionDenied{
		FileName: name,
	}
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("%s is not existed", e.Fname)
}

func (e *ErrFileParseFailed) Error() string {
	return fmt.Sprintf("parse %s failed: %s", e.Fname, e.Err)
}

func (e *ErrFlagFormat) Error() string {
	return fmt.Sprintf("flag: %s format error, please run '%s --help'", e.Flag, e.Cmd)
}

func (e *ErrCmdNotExist) Error() string {
	return fmt.Sprintf("command: %s not exist", e.Cmd)
}

func (e *ErrCmdNeedRoot) Error() string {
	return fmt.Sprintf("command: %s need run with sudo or root", e.Cmd)
}

func (e *ErrPermissionDenied) Error() string {
	return fmt.Sprintf("%s permission denied", e.FileName)
}
