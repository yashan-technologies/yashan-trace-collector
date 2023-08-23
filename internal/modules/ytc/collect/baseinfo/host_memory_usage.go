package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
)

func (b *BaseCollecter) hostMemoryUsage() (err error) {
	hostMemoryUsage := datadef.YTCItem{
		Name:     datadef.BASE_HOST_MEMORY_USAGE,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostMemoryUsage)

	log := log.Module.M(datadef.BASE_HOST_MEMORY_USAGE)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_MEMORY_USAGE)
	if err != nil {
		log.Errorf("failed to gert host memory usage info, err: %s", err.Error())
		hostMemoryUsage.Error = err.Error()
		return
	}
	hostMemoryUsage.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostMemoryUsage.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}
