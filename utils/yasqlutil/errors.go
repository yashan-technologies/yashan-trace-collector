package yasqlutil

import (
	"errors"
)

var (
	ErrInvalidSelect        = NewYasqlError("cannot select nil")
	ErrNoColumns            = NewYasqlError("cannot select no columns")
	ErrTableNotSet          = NewYasqlError("table not set")
	ErrInvalidDest          = NewYasqlError("type of dest should be pointer")
	ErrConnect              = NewYasqlError("connect failed")
	ErrDBNotOpen            = NewYasqlError("the database is not open")
	ErrYasqlHomeNotExist    = NewYasqlError("path: yasql-home not exist")
	ErrInvalidUserPwd       = NewYasqlError("invalid username/password, login denied")
	ErrLackLoginAuth        = NewYasqlError("user lacks CREATE SESSION privilege; logon denied")
	ErrNoDBUser             = NewYasqlError("user does not exist")
	ErrDBUserLackPrivileges = NewYasqlError("user lacks privileges")
	ErrInvalidUserPwdValue  = NewYasqlError("invalid username/password")
	ErrRecordNotFound       = NewYasqlError("record not found")
)

type yasqlError struct {
	msg string
}

func (e *yasqlError) Error() string {
	return e.msg
}

func NewYasqlError(message string) *yasqlError {
	return &yasqlError{msg: message}
}

func IsYasqlError(err error) bool {
	var e *yasqlError
	return errors.As(err, &e)
}
