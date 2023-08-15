package baseinfo

import (
	"ytc/defs/bashdef"
	"ytc/defs/errdef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/osutil"
	"ytc/utils/userutil"

	"git.yasdb.com/go/yasutil/execer"
)

func (b *BaseCollecter) CheckSarAccess() error {
	cmd := []string{
		"-c",
		bashdef.CMD_SAR,
		"-V",
	}
	exe := execer.NewExecer(log.Module)
	ret, stdout, stderr := exe.Exec(bashdef.CMD_BASH, cmd...)
	if ret != 0 || len(stderr) != 0 || len(stdout) == 0 {
		return errdef.NewErrCmdNotExist(bashdef.CMD_SAR)
	}
	return nil
}

func (b *BaseCollecter) CheckFireWallAssess() error {
	release, err := osutil.GetOsRelease()
	if err != nil {
		log.Module.Warnf("get os release err: %s", err)
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

func (b *BaseCollecter) checkYasdbEnv() error {
	env := yasdb.YasdbEnv{
		YasdbHome:     b.YasdbHome,
		YasdbData:     b.YasdbData,
		YasdbUser:     b.YasdbUser,
		YasdbPassword: b.YasdbPassword,
	}
	if err := env.ValidYasdbUserAndPwd(); err != nil {
		return err
	}
	return nil
}
