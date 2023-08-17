package collect

import (
	"ytc/defs/collecttypedef"
	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/internal/modules/ytc/collect/diagnosis"
)

type TypedCollecter interface {
	CheckAccess(yasdbValidate error) []data.NoAccessRes
	CollectFunc(item []string) map[string]func() error
	CollectedItem(noAccess []data.NoAccessRes) []string
	Type() string
	Start(packageDir string) error
	Finish() *data.YtcModule
}

func NewTypedCollecter(t string, collectParam *collecttypedef.CollectParam) (TypedCollecter, error) {
	switch t {
	case collecttypedef.TYPE_BASE:
		return baseinfo.NewBaseCollecter(collectParam), nil
	case collecttypedef.TYPE_DIAG:
		return diagnosis.NewDiagCollecter(collectParam), nil
	case collecttypedef.TYPE_PREF:
	}
	return nil, collecttypedef.ErrKnownType
}
