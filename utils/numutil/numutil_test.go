package numutil_test

import (
	"testing"

	"ytc/utils/numutil"
)

func TestTruncateFloat64(t *testing.T) {
	src := 3.141592653
	dest := 3.14
	result := numutil.TruncateFloat64(src, 2)
	if result != dest {
		t.Fatalf("result %f != %f", result, dest)
	}

	dest = 3.1416
	result = numutil.TruncateFloat64(src, 4)
	if result != dest {
		t.Fatalf("result %f != %f", result, dest)
	}
}
