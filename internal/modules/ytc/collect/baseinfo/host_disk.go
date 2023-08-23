package baseinfo

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"github.com/shirou/gopsutil/disk"
)

type DiskUsage struct {
	Device       string
	MountOptions string
	disk.UsageStat
}

func (b *BaseCollecter) getHostDiskInfo() (err error) {
	hostDiskInfo := datadef.YTCItem{Name: datadef.BASE_HOST_DISK}
	defer b.fillResult(&hostDiskInfo)

	log := log.Module.M(datadef.BASE_HOST_DISK)
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Errorf("failed to get host disk info, err: %s", err.Error())
		hostDiskInfo.Error = err.Error()
		return
	}
	var usages []DiskUsage
	for _, partition := range partitions {
		var usageStat *disk.UsageStat
		usageStat, err = disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Errorf("failed to get disk usage info, err: %s", err.Error())
			hostDiskInfo.Error = err.Error()
			return
		}
		usage := DiskUsage{
			Device:       partition.Device,
			MountOptions: partition.Opts,
			UsageStat:    *usageStat,
		}
		usages = append(usages, usage)
	}
	hostDiskInfo.Details = usages
	return
}
