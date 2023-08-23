package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"github.com/shirou/gopsutil/mem"
)

func (b *BaseCollecter) hostMemoryInfo() (err error) {
	hostMemoryInfo := datadef.YTCItem{Name: datadef.BASE_HOST_MEMORY}
	defer b.fillResult(&hostMemoryInfo)

	log := log.Module.M(datadef.BASE_HOST_MEMORY)
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("failed to get host memory info: %s", err.Error())
		hostMemoryInfo.Error = err.Error()
		return
	}
	hostMemoryInfo.Details = memInfo
	return
}
