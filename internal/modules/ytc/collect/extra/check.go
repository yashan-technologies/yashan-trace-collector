package extra

import (
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/utils/fileutil"
)

func (b *ExtraCollecter) checkExtraCollect() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.DIAG_HOST_DMESG
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

func (d *ExtraCollecter) CheckFunc() map[string]func() *ytccollectcommons.NoAccessRes {
	return map[string]func() *ytccollectcommons.NoAccessRes{
		datadef.EXTRA_FILE_COLLECT: d.checkExtraCollect,
	}
}
