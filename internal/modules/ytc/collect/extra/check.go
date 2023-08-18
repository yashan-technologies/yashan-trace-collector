package extra

import (
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/utils/fileutil"
)

func (b *ExtraCollecter) checkExtraCollect() *data.NoAccessRes {
	noAccess := new(data.NoAccessRes)
	noAccess.ModuleItem = data.DIAG_HOST_DMESG
	for _, path := range b.Include {
		if err := fileutil.CheckAccess(path); err != nil {
			tips, desc := ytccollectcommons.PathErrDescAndTips(path, err)
			noAccess.Description = tips
			noAccess.Tips = desc
			return noAccess
		}
	}
	// no need to check exclude
	return nil

}

func (d *ExtraCollecter) CheckFunc() map[string]func() *data.NoAccessRes {
	return map[string]func() *data.NoAccessRes{
		data.EXTRA_FILE_COLLECT: d.checkExtraCollect,
	}
}
