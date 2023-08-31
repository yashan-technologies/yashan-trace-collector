package diagnosis

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/yasqlutil"
)

func (b *DiagCollecter) getYasdbInstanceStatus() (err error) {
	yasdbInstanceStatusItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_INSTANCE_STATUS}
	defer b.fillResult(&yasdbInstanceStatusItem)

	log := log.Module.M(datadef.DIAG_YASDB_INSTANCE_STATUS)
	if b.notConnectDB {
		err = fmt.Errorf("connect failed, skip")
		yasdbInstanceStatusItem.Error = err.Error()
		yasdbInstanceStatusItem.Description = datadef.GenSkipCollectDatabaseInfoDesc()
		log.Error(err)
		return
	}
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	data, err := yasdb.QueryInstance(tx)
	if err != nil {
		log.Error(err)
		yasdbInstanceStatusItem.Error = err.Error()
		yasdbInstanceStatusItem.Description = datadef.GenDefaultDesc()
		return
	}
	yasdbInstanceStatusItem.Details = data
	return
}
