package userutil_test

import (
	"testing"
	"ytc/utils/userutil"

	"git.yasdb.com/go/yaslog"
)

func TestCheckSudovn(t *testing.T) {
	err := userutil.CheckSudovn(yaslog.NewDefaultConsoleLogger())
	if err != nil {
		t.Fatal(err)
	}
}
