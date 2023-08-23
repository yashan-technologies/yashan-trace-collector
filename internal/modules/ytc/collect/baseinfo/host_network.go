package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"github.com/shirou/gopsutil/net"
)

func (b *BaseCollecter) getHostNetworkInfo() (err error) {
	hostNetInfo := datadef.YTCItem{Name: datadef.BASE_HOST_NETWORK}
	defer b.fillResult(&hostNetInfo)

	log := log.Module.M(datadef.BASE_HOST_NETWORK)
	netInfo, err := net.Interfaces()
	if err != nil {
		log.Errorf("failed to get host network info, err: %s", err.Error())
		hostNetInfo.Error = err.Error()
		return
	}
	hostNetInfo.Details = netInfo
	return
}
