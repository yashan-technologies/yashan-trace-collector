package regexdef_test

import (
	"strconv"
	"testing"

	"ytc/defs/regexdef"
)

func TestRelative(t *testing.T) {
	cases := []struct {
		Name   string
		Path   string
		Expect bool
	}{
		{
			Name:   "相对路径./",
			Path:   "./",
			Expect: true,
		},
		{
			Name:   "相对路径../",
			Path:   "../",
			Expect: true,
		},
		{
			Name:   "相对路径只有字母",
			Path:   "a",
			Expect: true,
		},
		{
			Name:   "相对路径./../",
			Path:   "./../",
			Expect: true,
		},
		{
			Name:   "相对路径多级",
			Path:   "./asd/ccc",
			Expect: true,
		},
		{
			Name:   "相对路径多级字母开头",
			Path:   "a/asd/ccc",
			Expect: true,
		},
		{
			Name:   "据对路径含空格",
			Path:   "./ aaa/ccc",
			Expect: false,
		},
		{
			Name:   "据对路径含空格",
			Path:   "a/ aaa/ccc",
			Expect: false,
		},
		{
			Name:   "绝对路径",
			Path:   "/sdfas",
			Expect: true,
		},
		{
			Name:   "绝对路径含空格",
			Path:   "/sdf as",
			Expect: false,
		},
		{
			Name:   "据对路径多级",
			Path:   "/aaa/ccc/",
			Expect: true,
		},
		{
			Name:   "据对路径多级含空格",
			Path:   "/aaa/cc c/",
			Expect: false,
		},
		{
			Name:   "特殊书字符-_",
			Path:   "sdfas—_dgs",
			Expect: true,
		},
		{
			Name:   "绝对路径特殊书字符-_?",
			Path:   "/sdfas—_dgs/dfsasd?/",
			Expect: true,
		},
		{
			Name:   "绝对路径.",
			Path:   "/home/yashan/dsasf.1.1.a/build",
			Expect: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			res := regexdef.PATH_REGEX.MatchString(c.Path)
			if res != c.Expect {
				t.Fatalf("%s expect %s get %s", c.Path, strconv.FormatBool(c.Expect), strconv.FormatBool(res))
			}
		})
	}
}
