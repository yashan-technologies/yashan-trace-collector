package ytcctlhandler

import (
	"ytc/defs/collecttypedef"
	ytccollect "ytc/internal/modules/ytc/collect"
	"ytc/internal/modules/ytc/collect/data"
)

type CollecterHandler struct {
	Collecters    []ytccollect.TypedCollecter
	CollectResult *data.YtcReport
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
	return &CollecterHandler{
		Collecters:    typedCollecter,
		CollectResult: data.NewYtcReport(collectParam),
		Types:         types,
	}, nil
}
