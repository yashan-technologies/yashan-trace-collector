package baseinfo

import (
	"fmt"
	"path"

	"ytc/defs/bashdef"
	"ytc/defs/errdef"
	"ytc/defs/runtimedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/fileutil"
	"ytc/utils/osutil"
	"ytc/utils/userutil"
)

var (
	_os_sar_install_tips = map[string]string{
		osutil.CENTOS_ID: _tips_yum_base_host_load_status,
		osutil.UBUNTU_ID: _tips_apt_base_host_load_status,
		osutil.KYLIN_ID:  _tips_yum_base_host_load_status,
	}
)

func (b *BaseCollecter) CheckSarAccess() error {
	cmd := []string{
		"-c",
		fmt.Sprintf("%s -V", bashdef.CMD_SAR),
	}
	exe := execerutil.NewExecer(log.Module)
	ret, _, _ := exe.Exec(bashdef.CMD_BASH, cmd...)
	if ret != 0 {
		return errdef.NewErrCmdNotExist(bashdef.CMD_SAR)
	}
	return nil
}

func (b *BaseCollecter) CheckFireWallAccess() error {
	release := runtimedef.GetOSRelease()
	if release.Id != osutil.UBUNTU_ID {
		return nil
	}
	// ubuntu only for root
	if !userutil.IsCurrentUserRoot() {
		return errdef.NewErrCmdNeedRoot(bashdef.CMD_UFW)
	}
	return nil
}

func (b *BaseCollecter) checkYasdbVersion() *ytccollectcommons.NoAccessRes {
	yasdb := path.Join(b.YasdbHome, ytccollectcommons.BIN, ytccollectcommons.YASDB)
	err := fileutil.CheckAccess(yasdb)
	if err == nil {
		return nil
	}
	desc, tips := ytccollectcommons.PathErrDescAndTips(yasdb, err)
	return &ytccollectcommons.NoAccessRes{
		ModuleItem:  datadef.BASE_YASDB_VERION,
		Description: desc,
		Tips:        tips,
	}
}

func (b *BaseCollecter) checkYasdbParameter() (noAccess *ytccollectcommons.NoAccessRes) {
	noAccess = new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.BASE_YASDB_PARAMETER
	yasql := path.Join(b.YasdbHome, ytccollectcommons.BIN, ytccollectcommons.YASQL)
	ini := path.Join(b.YasdbData, ytccollectcommons.CONFIG, ytccollectcommons.YASDB_INI)
	iniErr := fileutil.CheckAccess(ini)
	yasqlErr := fileutil.CheckAccess(yasql)
	if yasqlErr != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(yasql, yasqlErr)
		if iniErr == nil {
			noAccess.ForceCollect = true
			ytccollectcommons.FillDescTips(noAccess, desc, fmt.Sprintf(ytccollectcommons.DEFAULT_PARAMETER_TIPS, ini))
			return
		}
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return
	}
	if b.yasdbValidateErr != nil {
		b.notConnectDB = true
		desc, tips := ytccollectcommons.YasErrDescAndTips(b.yasdbValidateErr)
		if iniErr == nil {
			noAccess.ForceCollect = true
			ytccollectcommons.FillDescTips(noAccess, desc, fmt.Sprintf(ytccollectcommons.DEFAULT_PARAMETER_TIPS, ini))
			return
		}
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return
	}
	return nil
}

func (b *BaseCollecter) checkFireWall() *ytccollectcommons.NoAccessRes {
	if err := b.CheckFireWallAccess(); err != nil {
		tips := err.Error()
		if err := userutil.CheckSudovn(log.Module); err != nil {
			tips = ytccollectcommons.CheckSudoTips(err)
		}
		return &ytccollectcommons.NoAccessRes{
			ModuleItem:  datadef.BASE_HOST_FIREWALLD,
			Description: err.Error(),
			Tips:        tips,
		}
	}
	return nil
}

func (b *BaseCollecter) checkNetworkIo() *ytccollectcommons.NoAccessRes {
	return b.checkSarWithItem(datadef.BASE_HOST_NETWORK_IO)
}

func (b *BaseCollecter) checkDiskIo() *ytccollectcommons.NoAccessRes {
	return b.checkSarWithItem(datadef.BASE_HOST_DISK_IO)

}

func (b *BaseCollecter) checkMemoryUsage() *ytccollectcommons.NoAccessRes {
	return b.checkSarWithItem(datadef.BASE_HOST_MEMORY_USAGE)

}

func (b *BaseCollecter) checkCpuUsage() *ytccollectcommons.NoAccessRes {
	return b.checkSarWithItem(datadef.BASE_HOST_CPU_USAGE)

}

func (b *BaseCollecter) checkSarWithItem(item string) *ytccollectcommons.NoAccessRes {
	if err := b.CheckSarAccess(); err != nil {
		os := runtimedef.GetOSRelease()
		noAccess := &ytccollectcommons.NoAccessRes{
			ModuleItem:   item,
			Description:  err.Error(),
			Tips:         fmt.Sprintf(_tips_sar_not_exist, _os_sar_install_tips[os.Id]),
			ForceCollect: true,
		}
		return noAccess
	}
	return nil
}

func (b *BaseCollecter) CheckFunc() map[string]checkFunc {
	return map[string]checkFunc{
		datadef.BASE_YASDB_VERION:      b.checkYasdbVersion,
		datadef.BASE_YASDB_PARAMETER:   b.checkYasdbParameter,
		datadef.BASE_HOST_FIREWALLD:    b.checkFireWall,
		datadef.BASE_HOST_NETWORK_IO:   b.checkNetworkIo,
		datadef.BASE_HOST_CPU_USAGE:    b.checkCpuUsage,
		datadef.BASE_HOST_DISK_IO:      b.checkDiskIo,
		datadef.BASE_HOST_MEMORY_USAGE: b.checkMemoryUsage,
	}
}
