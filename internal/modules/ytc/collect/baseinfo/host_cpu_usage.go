package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
)

func (b *BaseCollecter) hostCPUUsage() (err error) {
	hostCPUUsage := datadef.YTCItem{
		Name:     datadef.BASE_HOST_CPU_USAGE,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostCPUUsage)

	log := log.Module.M(datadef.BASE_HOST_CPU_USAGE)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_CPU_USAGE)
	if err != nil {
		log.Error("failed to get host cpu usage info, err: %s", err.Error())
		hostCPUUsage.Error = err.Error()
		return
	}
	hostCPUUsage.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostCPUUsage.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}
