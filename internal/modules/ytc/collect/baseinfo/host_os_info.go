package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"github.com/shirou/gopsutil/host"
)

func (b *BaseCollecter) getHostOSInfo() (err error) {
	hostBaseInfoItem := datadef.YTCItem{Name: datadef.BASE_HOST_OS_INFO}
	defer b.fillResult(&hostBaseInfoItem)

	log := log.Module.M(datadef.BASE_HOST_OS_INFO)
	hostInfo, err := host.Info()
	if err != nil {
		log.Errorf("failed to get host os info, err: %s", err.Error())
		hostBaseInfoItem.Error = err.Error()
		return
	}
	hostBaseInfoItem.Details = hostInfo
	return
}
