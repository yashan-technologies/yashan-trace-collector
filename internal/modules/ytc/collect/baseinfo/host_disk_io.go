package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
)

func (b *BaseCollecter) hostDiskIO() (err error) {
	hostDiskIO := datadef.YTCItem{
		Name:     datadef.BASE_HOST_DISK_IO,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostDiskIO)

	log := log.Module.M(datadef.BASE_HOST_DISK_IO)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_DISK_IO)
	if err != nil {
		log.Error("failed to get host disk IO info, err: %s", err.Error())
		hostDiskIO.Error = err.Error()
		return
	}
	hostDiskIO.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostDiskIO.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}
