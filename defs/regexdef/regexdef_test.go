package regexdef_test

import (
	"testing"
	"ytc/defs/regexdef"
)

func TestPath(t *testing.T) {
	cases := []struct {
		Name string
		Path string
	}{
		{
			Name: "根",
			Path: "/",
		},
		{
			Name: "一级",
			Path: "/car-c_a",
		},
		{
			Name: "二级",
			Path: "/car-c_a",
		},
		{
			Name: "相对",
			Path: "car",
		},
		{
			Name: "相对二级",
			Path: "car-c_a/car-c_a",
		},
		{
			Name: "相对./",
			Path: "./car-c_a/car-c_a",
		},
        {
			Name: "相对./",
			Path: "./car-c a/car-c*a",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			res := regexdef.PATH_REGEX.Match([]byte(c.Path))
			t.Logf("path: %s ,test res :%v", c.Path, res)
		})
	}
}
