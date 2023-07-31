package errdef

import "fmt"

type ErrFileNotFound struct {
	Fname string
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("%s is not existed", e.Fname)
}

type ErrFileParseFailed struct {
	Fname string
	Err   error
}

func (e *ErrFileParseFailed) Error() string {
	return fmt.Sprintf("parse %s failed: %s", e.Fname, e.Err)
}
