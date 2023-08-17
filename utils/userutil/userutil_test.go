package userutil_test

import (
	"strings"
	"testing"
	"ytc/utils/userutil"

	"git.yasdb.com/go/yaslog"
)

func TestCheckSudovn(t *testing.T) {
	err := userutil.CheckSudovn(yaslog.NewDefaultConsoleLogger())
	if err != nil && !strings.Contains(err.Error(), "a password is required") {
		t.Fatal(err)
	}
}
