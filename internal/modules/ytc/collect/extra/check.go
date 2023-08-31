package extra

import (
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/utils/fileutil"
)

func (b *ExtraCollecter) checkExtraCollect() *ytccollectcommons.NoAccessRes {
	noAccess := new(ytccollectcommons.NoAccessRes)
	noAccess.ModuleItem = datadef.EXTRA_FILE_COLLECT
	for _, path := range b.Include {
		res, err := fileutil.CheckDirAccess(path, b.genExcludeMap())
		if err != nil {
			desc, tips := ytccollectcommons.PathErrDescAndTips(path, err)
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		if len(res) != 0 {
			desc, tips := ytccollectcommons.FilesErrDescAndTips(res)
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			noAccess.ForceCollect = true
			return noAccess
		}
	}
	// no need to check exclude
	return nil

}

func (b *ExtraCollecter) CheckFunc() map[string]func() *ytccollectcommons.NoAccessRes {
	return map[string]func() *ytccollectcommons.NoAccessRes{
		datadef.EXTRA_FILE_COLLECT: b.checkExtraCollect,
	}
}
