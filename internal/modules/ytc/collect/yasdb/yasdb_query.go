package yasdb

import (
	"fmt"
	"ytc/utils/yasqlutil"
)

type VParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VInstance struct {
	Status string `json:"status"`
}

type VDatabase struct {
	Status   string `json:"status"`
	OpenMode string `json:"openMode"`
}

type ParameterName string

const (
	PM_DIAGNOSTIC_DEST   ParameterName = "DIAGNOSTIC_DEST"
	PM_RUN_LOG_FILE_PATH ParameterName = "RUN_LOG_FILE_PATH"
)

const (
	QueryYasdbAllParameter    = "select name,value from v$parameter where value is not null"
	QueryYasdbInstanceStatus  = "select status from v$instance;"
	QueryYasdbDatabaseStatus  = "select status,open_mode as openMode from v$database"
	QueryYasdbParameterByName = "select name,value from v$parameter where name='%s'"
)

var (
	_databaseSelecter = &yasqlutil.SelectRaw{
		RawSql: QueryYasdbDatabaseStatus,
	}
	_instanceSelecter = &yasqlutil.SelectRaw{
		RawSql: QueryYasdbInstanceStatus,
	}
)

func QueryParameter(tx *yasqlutil.Yasql, item ParameterName) (string, error) {
	tmp := &yasqlutil.SelectRaw{
		RawSql: fmt.Sprintf(QueryYasdbParameterByName, item),
	}
	pv := make([]*VParameter, 0)
	err := tx.SelectRaw(tmp).Find(&pv).Error()
	if err != nil {
		return "", err
	}
	if len(pv) == 0 {
		return "", yasqlutil.ErrRecordNotFound
	}
	return pv[0].Value, nil
}

func QueryAllParameter(tx *yasqlutil.Yasql) ([]*VParameter, error) {
	tmp := &yasqlutil.SelectRaw{
		RawSql: QueryYasdbAllParameter,
	}
	pv := make([]*VParameter, 0)
	err := tx.SelectRaw(tmp).Find(&pv).Error()
	return pv, err
}

func QueryDatabase(tx *yasqlutil.Yasql) (*VDatabase, error) {
	infos := make([]*VDatabase, 0)
	err := tx.SelectRaw(_databaseSelecter).Find(&infos).Error()
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, yasqlutil.ErrRecordNotFound
	}
	return infos[0], nil
}

func QueryInstance(tx *yasqlutil.Yasql) (*VInstance, error) {
	infos := make([]*VInstance, 0)
	err := tx.SelectRaw(_instanceSelecter).Find(&infos).Error()
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, yasqlutil.ErrRecordNotFound
	}
	return infos[0], nil
}
