package collect

import (
	"ytc/defs/collecttypedef"
	"ytc/internal/modules/ytc/collect/baseinfo"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/performance"
)

type TypedCollecter interface {
	CheckAccess(yasdbValidate error) []ytccollectcommons.NoAccessRes
	CollectFunc(item []string) map[string]func() error
	ItemsToCollect(noAccess []ytccollectcommons.NoAccessRes) []string
	Type() string
	PreCollect(packageDir string) error
	CollectOK() *datadef.YTCModule
}

func NewTypedCollecter(t string, collectParam *collecttypedef.CollectParam) (TypedCollecter, error) {
	switch t {
	case collecttypedef.TYPE_BASE:
		return baseinfo.NewBaseCollecter(collectParam), nil
	case collecttypedef.TYPE_DIAG:
		return diagnosis.NewDiagCollecter(collectParam), nil
	case collecttypedef.TYPE_PERF:
		return performance.NewPerfCollecter(collectParam), nil
	}
	return nil, collecttypedef.ErrKnownType
}
