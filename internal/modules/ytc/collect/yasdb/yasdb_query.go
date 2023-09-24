package yasdb

import (
	"errors"
	"fmt"
	"strings"

	"ytc/utils/yasqlutil"
)

type ParameterName string

const (
	PM_DIAGNOSTIC_DEST      ParameterName = "DIAGNOSTIC_DEST"
	PM_RUN_LOG_FILE_PATH    ParameterName = "RUN_LOG_FILE_PATH"
	SLOW_LOG_FILE_PATH      ParameterName = "SLOW_LOG_FILE_PATH"
	ENABLE_SLOW_LOG         ParameterName = "ENABLE_SLOW_LOG"
	SLOW_LOG_TIME_THRESHOLD ParameterName = "SLOW_LOG_TIME_THRESHOLD"
	SLOW_LOG_SQL_MAX_LEN    ParameterName = "SLOW_LOG_SQL_MAX_LEN"
	SLOW_LOG_OUTPUT         ParameterName = "SLOW_LOG_OUTPUT"
)

const (
	QUERY_YASDB_ALL_PARAMETER     = "select name,value from v$parameter where value is not null"
	QUERY_YASDB_INSTANCE_STATUS   = "select status from v$instance;"
	QUERY_YASDB_DATABASE_STATUS   = "select status,open_mode as openMode from v$database"
	QUERY_YASDB_PARAMETER_BY_NAME = "select name,value from v$parameter where name='%s'"
)

const (
	OPEN_MODE_READ_WRITE OpenMode = "READ_WRITE"
	OPEN_MODE_READ_ONLY  OpenMode = "READ_ONLY"
	OPEN_MODE_MOUNTED    OpenMode = "MOUNTED"
)

var (
	_databaseSelecter = &yasqlutil.SelectRaw{
		RawSql: QUERY_YASDB_DATABASE_STATUS,
	}
	_instanceSelecter = &yasqlutil.SelectRaw{
		RawSql: QUERY_YASDB_INSTANCE_STATUS,
	}

	_wrmDatabaseInstanceSelecter = &yasqlutil.Select{
		Table:   "sys.wrm$_database_instance",
		Columns: []string{"DBID", "INSTANCE_NUMBER"},
		ColTypes: map[string]string{
			"DBID":            "int64",
			"INSTANCE_NUMBER": "int64",
		},
	}

	_wrmSnapsotSelecter = &yasqlutil.Select{
		Table:   "sys.wrm$_snapshot",
		Columns: []string{"SNAP_ID", "DBID", "BEGIN_INTERVAL_TIME"},
		ColTypes: map[string]string{
			"SNAP_ID": "int64",
			"DBID":    "int64",
		},
	}

	_SlowLogSelector = &yasqlutil.Select{
		Table: "sys.SLOW_LOG$",
		Columns: []string{
			"DATABASE_NAME AS dbName",
			"USER_NAME as userName",
			"START_TIME AS startTime",
			"USER_HOST as userHost",
			"round(QUERY_TIME / 1000, 3) as queryTime",
			"ROWS_SENT as rowsSent",
			"SQL_ID as sqlID",
		},
		ColTypes: map[string]string{
			"ROWSSENT":  "int64",
			"QUERYTIME": "float64",
		},
	}

	_SqlTextSelector = &yasqlutil.Select{
		Table: "SLOW_LOG$",
		Columns: []string{
			"SQL_TEXT as sqlText",
		},
	}
)

var ErrNoSatisfiedSnapshot = errors.New("no snapshot id satisfied")

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

type WrmDatabaseInstance struct {
	DBID           int64 `json:"DBID"`
	InstanceNumber int64 `json:"INSTANCE_NUMBER"`
}

type WrmSnapshot struct {
	SnapID            int64  `json:"SNAP_ID"`
	DBID              int64  `json:"DBID"`
	BeginIntervalTime string `json:"BEGIN_INTERVAL_TIME"`
}

type SlowLog struct {
	DBName         string  `json:"dbName"`    // 数据库名称
	UserName       string  `json:"userName"`  // 用户名
	StartTime      string  `json:"startTime"` // 日志记录时的时间，非执行开始时间
	UserHost       string  `json:"userHost"`  // 连接IP
	QueryTime      float64 `json:"queryTime"` // 执行时间(ms)
	RowsSent       int64   `json:"rowsSent"`  // 查询返回的行数
	SQLID          string  `json:"sqlID"`     // sql id
	SQLText        string  `json:"sqlText"`   // sql语句
	StartTimestamp int64   `json:"-"`         // 日志记录时的时间，时间戳形式
}

type OpenMode string

func QueryParameter(tx *yasqlutil.Yasql, item ParameterName) (string, error) {
	tmp := &yasqlutil.SelectRaw{
		RawSql: fmt.Sprintf(QUERY_YASDB_PARAMETER_BY_NAME, item),
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
		RawSql: QUERY_YASDB_ALL_PARAMETER,
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

func QueryWrmDatabaseInstance(tx *yasqlutil.Yasql) (*WrmDatabaseInstance, error) {
	instances := make([]*WrmDatabaseInstance, 0)
	err := tx.Select(_wrmDatabaseInstanceSelecter).Find(&instances).Error()
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, yasqlutil.ErrRecordNotFound
	}
	return instances[0], nil
}

// return all snapshot between start end end
func QueryWrmSnapsot(tx *yasqlutil.Yasql, start string, end string) ([]*WrmSnapshot, error) {
	snaps := make([]*WrmSnapshot, 0)
	err := tx.Select(_wrmSnapsotSelecter).
		Where(fmt.Sprintf("BEGIN_INTERVAL_TIME >= TIMESTAMP('%s') and BEGIN_INTERVAL_TIME <= TIMESTAMP('%s')", start, end)).
		Find(&snaps).Error()
	if err != nil {
		return nil, err
	}
	if len(snaps) <= 1 {
		return nil, ErrNoSatisfiedSnapshot
	}
	return snaps, nil
}

// QuerySlowLog return all slow log between start and end
func QuerySlowLog(tx *yasqlutil.Yasql, start string, end string) ([]*SlowLog, error) {
	slows := make([]*SlowLog, 0)
	err := tx.Select(_SlowLogSelector).
		Where(fmt.Sprintf("START_TIME >= TIMESTAMP('%s') and START_TIME <= TIMESTAMP('%s')", start, end)).
		Find(&slows).Error()
	if err != nil {
		return nil, err
	}
	for _, s := range slows {
		if err := s.afterFind(tx); err != nil {
			return nil, err
		}
	}
	return slows, nil
}

func (s *SlowLog) afterFind(tx *yasqlutil.Yasql) error {
	newTx := yasqlutil.GetLocalInstance(tx.User, tx.Password, tx.YasqlHome, tx.YasdbData)
	slowlogItems := []*SlowLog{}
	if err := newTx.Select(_SqlTextSelector).Where("SQL_ID = ? and START_TIME = ? ", s.SQLID, s.StartTime).Find(&slowlogItems).Error(); err != nil {
		return err
	}
	texts := []string{}
	for _, slowlog := range slowlogItems {
		texts = append(texts, slowlog.SQLText)
	}
	s.SQLText = strings.Join(texts, "\n")
	return nil
}

func (d *VDatabase) IsDatabaseInReadWwiteMode() bool {
	return d.OpenMode == string(OPEN_MODE_READ_WRITE)
}
