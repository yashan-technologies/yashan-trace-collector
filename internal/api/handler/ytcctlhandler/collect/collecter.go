package ytcctlhandler

import (
	"ytc/defs/collecttypedef"
	ytccollect "ytc/internal/modules/ytc/collect"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/internal/modules/ytc/collect/extra"
)

type CollecterHandler struct {
	Collecters    []ytccollect.TypedCollecter
	CollectResult *data.YTCReport
	Types         map[string]struct{}
}

func NewCollecterHandler(types map[string]struct{}, collectParam *collecttypedef.CollectParam) (*CollecterHandler, error) {
	typedCollecter := make([]ytccollect.TypedCollecter, 0)
	for t := range types {
		c, err := ytccollect.NewTypedCollecter(t, collectParam)
		if err != nil {
			return nil, err
		}
		typedCollecter = append(typedCollecter, c)
	}
	if len(collectParam.Include) != 0 {
		typedCollecter = append(typedCollecter, extra.NewExtraCollecter(collectParam))
	}
	return &CollecterHandler{
		Collecters:    typedCollecter,
		CollectResult: data.NewYTCReport(collectParam),
		Types:         types,
	}, nil
}
