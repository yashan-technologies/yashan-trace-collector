package osutil_test

import (
	"testing"

	"ytc/utils/jsonutil"
	"ytc/utils/osutil"
)

func TestOsRelease(t *testing.T) {
	release, err := osutil.GetOsRelease()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(jsonutil.ToJSONString(release))
}
