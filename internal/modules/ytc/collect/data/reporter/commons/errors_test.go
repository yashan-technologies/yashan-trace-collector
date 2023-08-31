package commons_test

import (
	"testing"

	"ytc/internal/modules/ytc/collect/data/reporter/commons"
)

func TestErrInterfaceTypeNotMatch(t *testing.T) {
	err := &commons.ErrInterfaceTypeNotMatch{
		Key:     "YashanDB Version",
		Targets: []interface{}{"", 1, 1.1},
		Current: false,
	}
	t.Log(err)

	err = &commons.ErrInterfaceTypeNotMatch{
		Key:     "YashanDB Version",
		Targets: []interface{}{""},
		Current: false,
	}
	t.Log(err)
}
