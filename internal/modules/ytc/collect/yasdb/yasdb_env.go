package yasdb

import (
	"os"
	constdef "ytc/defs/constants"
	"ytc/defs/errdef"
	"ytc/utils/stringutil"
	"ytc/utils/yasqlutil"
)

type YasdbEnv struct {
	YasdbHome     string `json:"yasdbHome"`
	YasdbData     string `json:"yasdbData"`
	YasdbUser     string `json:"yasdbUser"`
	YasdbPassword string `json:"yasdbPassword"`
}

func (y *YasdbEnv) ValidYasdbHome() error {
	if stringutil.IsEmpty(y.YasdbHome) {
		return errdef.NewItemEmpty(constdef.YASDB_HOME)
	}
	if err := y.validatePath(y.YasdbHome); err != nil {
		return err
	}
	return nil
}

func (y *YasdbEnv) ValidYasdbData() error {
	if stringutil.IsEmpty(y.YasdbData) {
		return errdef.NewItemEmpty(constdef.YASDB_DATA)
	}
	if err := y.validatePath(y.YasdbData); err != nil {
		return err
	}
	return nil
}

func (y *YasdbEnv) ValidYasdbUser() error {
	if stringutil.IsEmpty(y.YasdbUser) {
		return errdef.NewItemEmpty(constdef.YASDB_USER)
	}
	return nil
}

func (y *YasdbEnv) ValidYasdbPassword() error {
	if stringutil.IsEmpty(y.YasdbPassword) {
		return errdef.NewItemEmpty(constdef.YASDB_DATA)
	}
	return nil
}

func (y *YasdbEnv) ValidYasdbUserAndPwd() error {
	if err := y.ValidYasdbHome(); err != nil {
		return err
	}
	if err := y.ValidYasdbData(); err != nil {
		return err
	}
	if err := y.ValidYasdbUser(); err != nil {
		return err
	}

	if err := y.ValidYasdbPassword(); err != nil {
		return err
	}
	tx := yasqlutil.GetLocalInstance(y.YasdbUser, y.YasdbPassword, y.YasdbHome, y.YasdbData)
	if err := tx.CheckPassword(); err != nil {
		return err
	}
	return nil
}

func (y *YasdbEnv) validatePath(path string) error {
	_, err := os.Stat(path)
	return err
}
