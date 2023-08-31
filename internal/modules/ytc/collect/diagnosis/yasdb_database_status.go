package diagnosis

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/yasqlutil"
)

func (b *DiagCollecter) getYasdbDatabaseStatus() (err error) {
	yasdbDatabaseStatusItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_DATABASE_STATUS}
	defer b.fillResult(&yasdbDatabaseStatusItem)

	log := log.Module.M(datadef.DIAG_YASDB_DATABASE_STATUS)
	if b.notConnectDB {
		err = fmt.Errorf("connect failed, skip")
		yasdbDatabaseStatusItem.Error = err.Error()
		yasdbDatabaseStatusItem.Description = datadef.GenSkipCollectDatabaseInfoDesc()
		log.Error(err)
		return
	}
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	data, err := yasdb.QueryDatabase(tx)
	if err != nil {
		log.Error(err)
		yasdbDatabaseStatusItem.Error = err.Error()
		yasdbDatabaseStatusItem.Description = datadef.GenDefaultDesc()
		return
	}
	yasdbDatabaseStatusItem.Details = data
	return
}
