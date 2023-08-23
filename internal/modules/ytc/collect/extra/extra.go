package extra

import (
	"path"

	"ytc/defs/collecttypedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"git.yasdb.com/go/yasutil/fs"
)

const (
	EXTRA_DIR_NAME = "extra"
)

var _packageDir = ""

var (
	ExtraChineseName = map[string]string{
		datadef.EXTRA_FILE_COLLECT: "额外文件收集",
	}
)

type ExtraCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *datadef.YTCModule
}

func NewExtraCollecter(collectParam *collecttypedef.CollectParam) *ExtraCollecter {
	return &ExtraCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &datadef.YTCModule{
			Module: collecttypedef.TYPE_EXTRA,
		},
	}
}

// [Interface Func]
func (b *ExtraCollecter) Type() string {
	return collecttypedef.TYPE_EXTRA
}

// [Interface Func]
func (b *ExtraCollecter) CheckAccess(yasdbValidate error) (noAccess []ytccollectcommons.NoAccessRes) {
	noAccess = make([]ytccollectcommons.NoAccessRes, 0)
	funcMap := b.CheckFunc()
	for item, fn := range funcMap {
		noAccessRes := fn()
		if noAccessRes != nil {
			log.Module.Debugf("item [%s] check asscess desc: %s tips %s", item, noAccessRes.Description, noAccessRes.Tips)
			noAccess = append(noAccess, *noAccessRes)
		}
	}
	return
}

// [Interface Func]
func (b *ExtraCollecter) CollectFunc(items []string) (res map[string]func() error) {
	res = make(map[string]func() error)
	itemFuncMap := b.itemFunc()
	for _, collectItem := range items {
		_, ok := itemFuncMap[collectItem]
		if !ok {
			log.Module.Errorf("get %s collect func err %s", collectItem)
			continue
		}
		res[collectItem] = itemFuncMap[collectItem]
	}
	return
}

// [Interface Func]
func (b *ExtraCollecter) ItemsToCollect(noAccess []ytccollectcommons.NoAccessRes) (res []string) {
	noMap := b.getNotAccessItem(noAccess)
	for item := range ExtraChineseName {
		if _, ok := noMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

func (b *ExtraCollecter) getNotAccessItem(noAccess []ytccollectcommons.NoAccessRes) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, noAccessRes := range noAccess {
		if noAccessRes.ForceCollect {
			continue
		}
		res[noAccessRes.ModuleItem] = struct{}{}
	}
	return
}

// [Interface Func]
func (b *ExtraCollecter) PreCollect(packageDir string) (err error) {
	b.setPackageDir(packageDir)
	err = fs.Mkdir(path.Join(packageDir, EXTRA_DIR_NAME))
	return
}

// [Interface Func]
func (b *ExtraCollecter) CollectOK() *datadef.YTCModule {
	return b.ModuleCollectRes
}

func (b *ExtraCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		datadef.EXTRA_FILE_COLLECT: b.collectExtraFile,
	}
}

func (b *ExtraCollecter) setPackageDir(packageDir string) {
	_packageDir = packageDir
}

func (b *ExtraCollecter) fillResult(data *datadef.YTCItem) {
	b.ModuleCollectRes.Set(data)
}
