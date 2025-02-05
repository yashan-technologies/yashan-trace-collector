package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"github.com/shirou/gopsutil/cpu"
)

func (b *BaseCollecter) getHostCPUInfo() (err error) {
	hostCpuInfo := datadef.YTCItem{Name: datadef.BASE_HOST_CPU}
	defer b.fillResult(&hostCpuInfo)

	log := log.Module.M(datadef.BASE_HOST_CPU)
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Errorf("failed to get host cpu info, err: %s", err.Error())
		hostCpuInfo.Error = err.Error()
		hostCpuInfo.Description = datadef.GenDefaultDesc()
		return
	}
	hostCpuInfo.Details = cpuInfo
	return
}
