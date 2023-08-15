package ytccollectcommons

import (
	"errors"
	"ytc/defs/errdef"
	"ytc/utils/yasqlutil"
)

func GenYasdbEnvErrTips(err error) (tips string) {
	var e *errdef.ItemEmpty
	if errors.As(err, &e) {
		tips = "Re-run 'ytcctl collect' and full yashandb user and yashandb password"
	}
	if yasqlutil.IsYasqlError(err) {
		if err == yasqlutil.ErrDBUserLackPrivileges {
			tips = "grant create session to current use yasdb user"
		}
		if err == yasqlutil.ErrConnect {
			tips = "check process fo yashandb"
		}
	}
	return
}
