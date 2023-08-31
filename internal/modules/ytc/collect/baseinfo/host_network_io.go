package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
)

func (b *BaseCollecter) getHostNetworkIO() (err error) {
	hostNetworkIO := datadef.YTCItem{
		Name:     datadef.BASE_HOST_NETWORK_IO,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostNetworkIO)

	log := log.Module.M(datadef.BASE_HOST_NETWORK_IO)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_NETWORK_IO)
	if err != nil {
		log.Errorf("failed to get host network IO info, err: %s", err.Error())
		hostNetworkIO.Error = err.Error()
		hostNetworkIO.Description = datadef.GenHostWorkloadDesc(err)
		return
	}
	hostNetworkIO.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostNetworkIO.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}
