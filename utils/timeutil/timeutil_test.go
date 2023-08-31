package timeutil_test

import (
	"testing"

	"ytc/defs/timedef"
	"ytc/utils/timeutil"
)

func TestGetTimeDivBySepa(t *testing.T) {
	cases := []struct {
		Name    string
		TimeStr string
		Spea    string
	}{
		{
			Name:    "1",
			TimeStr: "2023-01-21-02-03-03",
			Spea:    "-",
		},
		{
			Name:    "1",
			TimeStr: "2023-01-21-02-03",
			Spea:    "-",
		},
		{
			Name:    "1",
			TimeStr: "2023-01-21",
			Spea:    "-",
		},
		{
			Name:    "1",
			TimeStr: "2023-01",
			Spea:    "-",
		},
		{
			Name:    "1",
			TimeStr: "2023",
			Spea:    "-",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			time, err := timeutil.GetTimeDivBySepa(c.TimeStr, c.Spea)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(time.Format(timedef.TIME_FORMAT))
		})
	}

}
