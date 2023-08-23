package diagnosis

import (
	"ytc/defs/errdef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/processutil"
)

func (b *DiagCollecter) yasdbProcessStatus() (err error) {
	yasdbProcessStatusItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_PROCESS_STATUS}
	defer b.fillResult(&yasdbProcessStatusItem)

	log := log.Module.M(datadef.DIAG_YASDB_PROCESS_STATUS)
	processes, err := processutil.GetYasdbProcess(b.YasdbData)
	if err != nil {
		log.Error(err)
		yasdbProcessStatusItem.Error = err.Error()
		return
	}
	if len(processes) == 0 {
		err = errdef.NewErrYasdbProcessNotFound()
		log.Error(err)
		yasdbProcessStatusItem.Error = err.Error()
		return
	}
	proc := processes[0]
	if err = proc.FindBaseInfo(); err != nil {
		log.Error(err)
		yasdbProcessStatusItem.Error = err.Error()
		return
	}
	yasdbProcessStatusItem.Details = proc
	return
}
