package collect

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	ytcctlhandler "ytc/internal/api/handler/ytcctlhandler/collect"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/jsonutil"
	"ytc/utils/stringutil"
	"ytc/utils/terminalutil"
	"ytc/utils/timeutil"
)

const (
	ytctl_collect = "ytctl collect"
	f_type        = "--type"
	f_range       = "--range"
	f_start       = "--start"
	f_end         = "--end"
	f_output      = "--output"
)

type CollectGlobal struct {
	Type    string `name:"type"   short:"t" default:"base,diag" help:"The type of collection, choose many of (base|diag|perf) and split with ','."`
	Range   string `name:"range"  short:"r" help:"The time range of the collection, such as ''1M', '1d', '1h', '1m'. If <range> is given, <start> and <end> will be discard."`
	Start   string `name:"start"  short:"s" help:"The start datetime of the collection, such as 'yyyy-MM-dd', 'yyyy-MM-dd-hh', 'yyyy-MM-dd-hh-mm'"`
	End     string `name:"end"    short:"e" help:"The end timestamp of the collection, such as 'yyyy-MM-dd', 'yyyy-MM-dd-hh', 'yyyy-MM-dd-hh-mm',, default value is current datetime."`
	Output  string `name:"output" short:"o" help:"The output dir of the collection."`
	Include string `name:"include" help:"Files or directories that need to be additionally collected, it is absolute path and split with ',', such as '/tmp' or '/tmp,/root,/example.txt'."`
	Exclude string `name:"exclude" help:"Files or directories that no need to be additionally collected, it is absolute path and split with ',', such as '/tmp' or '/tmp,/root,/example.txt'."`
}

type CollectCmd struct {
	CollectGlobal
}

// [Interface Func]
func (c *CollectCmd) Run() error {
	if err := c.validate(); err != nil {
		return err
	}
	yasdbEnv, code := c.openYasdbCollectForm()
	if code == terminalutil.FormExitNotContinue {
		c.Quit()
		return nil
	}
	collectParam, err := c.genCollcterParam(yasdbEnv)
	if err != nil {
		log.Controller.Errorf("get collect info err %s", err.Error())
		return err
	}
	types, err := c.getTypes()
	if err != nil {
		return err
	}
	handler, err := ytcctlhandler.NewCollecterHandler(types, collectParam)
	if err != nil {
		return err
	}
	log.Controller.Debugf("from validate res :%s, ", jsonutil.ToJSONString(YasdbValidate))
	if err := handler.Collect(YasdbValidate); err != nil {
		log.Controller.Errorf(err.Error())
		fmt.Printf("Stopping Collect...\n\n")
	}
	return nil
}

func (c *CollectCmd) Quit() {
	fmt.Println("Quit Collect")
}

func (c *CollectCmd) genCollcterParam(env *yasdb.YasdbEnv) (*collecttypedef.CollectParam, error) {
	start, end, err := c.getStartAndEnd()
	if err != nil {
		return nil, err
	}
	return &collecttypedef.CollectParam{
		StartTime:     start,
		EndTime:       end,
		Output:        c.Output,
		YasdbHome:     env.YasdbHome,
		YasdbData:     env.YasdbData,
		YasdbUser:     env.YasdbUser,
		YasdbPassword: env.YasdbPassword,
		Include:       c.getExtraPath(c.Include),
		Exclude:       c.getExtraPath(c.Exclude),
	}, nil
}

func (c *CollectCmd) getStartAndEnd() (start time.Time, end time.Time, err error) {
	defer func() {
		end = end.Add(time.Minute)
	}()
	startegy := confdef.GetStrategyConf()
	defRange := startegy.Collect.GetRange()
	if !stringutil.IsEmpty(c.Range) {
		if err = c.validateRange(); err != nil {
			return
		}
		var r time.Duration
		r, err = timeutil.GetDuration(c.Range)
		if err != nil {
			return
		}
		end = time.Now()
		start = end.Add(-r)
		return
	}
	// start and end
	if err = c.validateStartAndEnd(); err != nil {
		return
	}
	if !stringutil.IsEmpty(c.Start) {
		start, err = timeutil.GetTimeDivBySepa(c.Start, stringutil.STR_HYPHEN)
		if err != nil {
			return
		}
		// only start
		if stringutil.IsEmpty(c.End) {
			end = start.Add(defRange)
			return
		}
		// both start end
		end, err = timeutil.GetTimeDivBySepa(c.End, stringutil.STR_HYPHEN)
		if err != nil {
			return
		}
		return
	}
	// only end
	if !stringutil.IsEmpty(c.End) {
		end, err = timeutil.GetTimeDivBySepa(c.End, stringutil.STR_HYPHEN)
		if err != nil {
			return
		}
		start = end.Add(-defRange)
		return
	}
	end = time.Now()
	start = end.Add(-defRange)
	return
}

func (c *CollectCmd) getTypes() (types map[string]struct{}, err error) {
	types = make(map[string]struct{})
	if err = c.validateType(); err != nil {
		return
	}
	fields := strings.Split(c.Type, stringutil.STR_COMMA)
	for _, f := range fields {
		types[f] = struct{}{}
	}
	return
}

func (c *CollectCmd) getExtraPath(value string) (res []string) {
	if stringutil.IsEmpty(value) {
		return
	}
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, stringutil.STR_COMMA)
	value = strings.TrimSuffix(value, stringutil.STR_COMMA)
	fields := strings.Split(value, stringutil.STR_COMMA)
	for _, field := range fields {
		res = append(res, filepath.Clean(strings.TrimSpace(field)))
	}
	return
}
