package collect

import (
	"os"
	"syscall"
	"testing"
)

func TestWrite(t *testing.T) {
	path := "/home/yashan/ccccc"
	d, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Log(err)
		}
		if err := os.MkdirAll(path, 0775); err != nil {
			t.Log(err)
            return
		}
	}

	res := d.Mode().Perm()&syscall.S_IWRITE == 0
	t.Log(res)
}
