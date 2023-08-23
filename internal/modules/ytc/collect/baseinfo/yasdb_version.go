package baseinfo

import (
	"fmt"
	"path"
	"strings"

	"ytc/defs/bashdef"
	"ytc/defs/errdef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yasutil/fs"
)

func (b *BaseCollecter) getYasdbVersion() (err error) {
	yasdbVersionItem := datadef.YTCItem{Name: datadef.BASE_YASDB_VERION}
	defer b.fillResult(&yasdbVersionItem)

	log := log.Module.M(datadef.BASE_YASDB_VERION)
	yasdbBinPath := path.Join(b.YasdbHome, "bin", bashdef.CMD_YASDB)
	if !fs.IsFileExist(yasdbBinPath) {
		err = &errdef.ErrFileNotFound{Fname: yasdbBinPath}
		log.Errorf("failed to get yashandb version, err: %s", err.Error())
		yasdbVersionItem.Error = err.Error()
		// TODO: 补充description信息
		return
	}
	execer := execerutil.NewExecer(log)
	env := []string{fmt.Sprintf("%s=%s", yasqlutil.LIB_KEY, path.Join(b.YasdbHome, yasqlutil.LIB_PATH))}
	ret, stdout, stderr := execer.EnvExec(env, yasdbBinPath, "-V")
	if ret != 0 {
		err = fmt.Errorf("failed to get yasdb version, err: %s", stderr)
		log.Error(err)
		yasdbVersionItem.Error = stderr
		return
	}
	yasdbVersionItem.Details = strings.TrimSpace(stdout)
	return
}
