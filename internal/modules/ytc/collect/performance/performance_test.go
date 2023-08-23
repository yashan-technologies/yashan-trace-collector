package performance

import (
	"testing"
	"time"

	"ytc/defs/collecttypedef"
	"ytc/defs/timedef"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/assert"
)

var (
	logLines = []string{
		"# TIME: 2023-08-21 16:47:29.963",
		"# USER_HOST: SYS@192.168.8.247",
		"# DB_NAME: ha",
		"# COST_EXECUTE_TIME: 31.071857",
		"# COST_OPTIMIZE_TIME: 0.000000",
		"# ROWS_SENT: 0",
		"# SQL_ID: 6pk27y53wjrqv",
		"SQL: insert into s values(2,'ss')",
		"   where name = 'aaa' and id = 1;",
		"# TIME: 2023-07-27 16:49:27.962",
		"# USER_HOST: SYS@192.168.8.247",
		"# DB_NAME: ha",
		"# COST_EXECUTE_TIME: 80.797355",
		"# COST_OPTIMIZE_TIME: 0.000000",
		"# ROWS_SENT: 1",
		"# SQL_ID: dnnqmx0r4yabd",
		"SQL: insert into s values(4,'qq')",
	}
)

func TestFilterSlowLog(t *testing.T) {
	start, _ := time.ParseInLocation(timedef.TIME_FORMAT, "2023-08-21 10:03:17", time.Local)
	end, _ := time.ParseInLocation(timedef.TIME_FORMAT, "2023-08-22 10:05:17", time.Local)
	p := PerfCollecter{
		CollectParam: &collecttypedef.CollectParam{
			StartTime: start,
			EndTime:   end,
		},
	}
	resLine := []string{}
	p.filterSlow(yaslog.NewDefaultConsoleLogger(), logLines, &resLine)
	a := assert.NewAssert(t)
	a.Equal(9, len(resLine))
}
