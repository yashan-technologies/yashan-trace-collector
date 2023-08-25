package errdef

import "errors"

var (
	ErrNoneCollectTtem = errors.New("no collection items will be collected, skip this collection")
)
