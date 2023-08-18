package baseinfo

import (
	"fmt"
	"path"

	"ytc/defs/bashdef"
	"ytc/defs/errdef"
	ytccommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/fileutil"
	"ytc/utils/osutil"
	"ytc/utils/userutil"
)

func (b *BaseCollecter) CheckSarAccess() error {
	cmd := []string{
		"-c",
		bashdef.CMD_SAR,
		"-V",
	}
	exe := execerutil.NewExecer(log.Module)
	ret, stdout, stderr := exe.Exec(bashdef.CMD_BASH, cmd...)
	if ret != 0 || len(stderr) != 0 || len(stdout) == 0 {
		return errdef.NewErrCmdNotExist(bashdef.CMD_SAR)
	}
	return nil
}

func (b *BaseCollecter) CheckFireWallAssess() error {
	release, err := osutil.GetOsRelease()
	if err != nil {
		log.Module.Warnf(ytccommons.GetOsReleaseErrDesc, err)
		return err
	}
	if release.Id != osutil.UBUNTU_ID {
		return nil
	}
	// ubuntu only for root
	if !userutil.IsCurrentUserRoot() {
		return errdef.NewErrCmdNeedRoot(bashdef.CMD_UFW)
	}
	return nil
}

func (b *BaseCollecter) checkYasdbVersion() *data.NoAccessRes {
	yasdb := path.Join(b.YasdbHome, ytccommons.BIN, ytccommons.YASDB)
	err := fileutil.CheckAccess(yasdb)
	if err == nil {
		return nil
	}
	desc, tips := ytccommons.PathErrDescAndTips(yasdb, err)
	return &data.NoAccessRes{
		ModuleItem:  data.BASE_YASDB_VERION,
		Description: desc,
		Tips:        tips,
	}
}

func (b *BaseCollecter) checkYasdbParameter() (noAccess *data.NoAccessRes) {
	noAccess = new(data.NoAccessRes)
	noAccess.ModuleItem = data.BASE_YASDB_PARAMTER
	yasql := path.Join(b.YasdbHome, ytccommons.BIN, ytccommons.YASQL)
	ini := path.Join(b.YasdbData, ytccommons.CONFIG, ytccommons.YASDB_INI)
	iniErr := fileutil.CheckAccess(ini)
	yasqlErr := fileutil.CheckAccess(yasql)
	if yasqlErr != nil {
		desc, tips := ytccommons.PathErrDescAndTips(yasql, yasqlErr)
		if iniErr == nil {
			noAccess.ForceCollect = true
			ytccommons.FullDescTips(noAccess, desc, fmt.Sprintf(ytccommons.DefaultParameterTips, ini))
			return
		}
		ytccommons.FullDescTips(noAccess, desc, tips)
		return
	}
	if b.yasdbValidateErr != nil {
		b.notConnectDB = true
		desc, tips := ytccommons.YasErrDescAndtips(b.yasdbValidateErr)
		if iniErr == nil {
			noAccess.ForceCollect = true
			ytccommons.FullDescTips(noAccess, desc, fmt.Sprintf(ytccommons.DefaultParameterTips, ini))
			return
		}
		ytccommons.FullDescTips(noAccess, desc, tips)
		return
	}
	return nil
}

func (b *BaseCollecter) checkFireWall() *data.NoAccessRes {
	if err := b.CheckFireWallAssess(); err != nil {
		tips := err.Error()
		if err := userutil.CheckSudovn(log.Module); err != nil {
			tips = ytccommons.CheckSudoTips(err)
		}
		return &data.NoAccessRes{
			ModuleItem:  data.BASE_HOST_FIREWALLD,
			Description: err.Error(),
			Tips:        tips,
		}
	}
	return nil
}

func (b *BaseCollecter) checkNetworkIo() *data.NoAccessRes {
	return b.checkSarWithItem(data.BASE_HOST_NETWORK_IO)
}

func (b *BaseCollecter) checkDiskIo() *data.NoAccessRes {
	return b.checkSarWithItem(data.BASE_HOST_DISK_IO)

}

func (b *BaseCollecter) checkMemoryUsage() *data.NoAccessRes {
	return b.checkSarWithItem(data.BASE_HOST_MEMORY_USAGE)

}

func (b *BaseCollecter) checkCpuUsage() *data.NoAccessRes {
	return b.checkSarWithItem(data.BASE_HOST_CPU_USAGE)

}

func (b *BaseCollecter) checkSarWithItem(item string) *data.NoAccessRes {
	if err := b.CheckSarAccess(); err != nil {
		os, osErr := osutil.GetOsRelease()
		if osErr != nil {
			log.Module.Errorf(ytccommons.GetOsReleaseErrDesc, err.Error())
		}
		var tips string
		if os.Id == osutil.UBUNTU_ID {
			tips = _tips_apt_base_host_load_status
		}
		if os.Id == osutil.CENTOS_ID {
			tips = _tips_yum_base_host_load_status
		}
		if os.Id == osutil.KYLIN_ID {
			tips = _tips_yum_base_host_load_status
		}
		noAccess := &data.NoAccessRes{
			ModuleItem:  item,
			Description: err.Error(),
			Tips:        tips,
		}
		return noAccess
	}
	return nil
}

func (b *BaseCollecter) CheckFunc() map[string]checkFunc {
	return map[string]checkFunc{
		data.BASE_YASDB_VERION:      b.checkYasdbVersion,
		data.BASE_YASDB_PARAMTER:    b.checkYasdbParameter,
		data.BASE_HOST_FIREWALLD:    b.checkFireWall,
		data.BASE_HOST_NETWORK_IO:   b.checkNetworkIo,
		data.BASE_HOST_CPU_USAGE:    b.checkCpuUsage,
		data.BASE_HOST_DISK_IO:      b.checkDiskIo,
		data.BASE_HOST_MEMORY_USAGE: b.checkMemoryUsage,
	}
}
