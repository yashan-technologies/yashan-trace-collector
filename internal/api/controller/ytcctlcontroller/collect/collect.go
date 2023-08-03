package collect

import (
	"fmt"
	"strings"
	"ytc/defs/confdef"
	collecthandler "ytc/internal/api/handler/ytcctlhandler/collect"
	"ytc/log"
	"ytc/utils/stringutil"
)

type CollectGlobal struct {
	Type       string `name:"type"   short:"t" default:"base,diag" help:"The type of collection, choose many of (base|diag|perf) and split with ','."`
	Range      string `name:"range"  short:"r" help:"The time range of the collection, such as '24h', '48h'. If <range> is given, <start> and <end> will be discard."`
	Start      int64  `name:"start"  short:"s" help:"The start timestamp of the collection."`
	End        int64  `name:"end"    short:"e" help:"The end timestamp of the collection, default value is current timestamp."`
	Output     string `name:"output" short:"o" help:"The output dir of the collection."`
	ReportType string `name:"report-type" help:"Type of report generated, choose one from (txt)."`
}

type CollectCmd struct {
	CollectGlobal
}

// [Interface Func]
func (c CollectCmd) Run() error {
	if stringutil.IsEmpty(c.Output) {
		c.Output = confdef.GetStrategyConf().Collect.Output
	}
	if stringutil.IsEmpty(c.ReportType) {
		c.ReportType = confdef.GetStrategyConf().Report.Type
	}
	types := make(map[string]struct{})
	for _, s := range strings.Split(c.Type, stringutil.STR_COMMA) {
		types[s] = struct{}{}
	}
	result, err := collecthandler.Demo(c.Output, c.ReportType, types)
	if err != nil {
		log.Controller.Errorf("collecthandler call demo failed: %s", err)
		return err
	}
	fmt.Println("collect finished, result locates at:", result)
	return nil
}
