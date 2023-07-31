package main

import (
	"ytc/commons/flags"
	"ytc/internal/api/controller/ytcctlcontroller/clean"
	"ytc/internal/api/controller/ytcctlcontroller/collect"
	"ytc/internal/api/controller/ytcctlcontroller/daemon"
	"ytc/internal/api/controller/ytcctlcontroller/report"
	"ytc/internal/api/controller/ytcctlcontroller/strategy"
	"ytc/internal/api/controller/ytcctlcontroller/yasdb"
)

type App struct {
	flags.Globals

	Collect collect.CollectCmd `cmd:"collect" name:"collect" help:"The collect command is used to gather trace data."`
	// TODO: remove hidden:"true" when commands are supported
	Daemon   daemon.DaemonCmd     `cmd:"daemon"   name:"daemon"   hidden:"true" help:"The daemon command is used to manage the life cycle of ytcd."`
	Strategy strategy.StrategyCmd `cmd:"strategy" name:"strategy" hidden:"true" help:"The strategy command is used to manage the collector strategy."`
	Clean    clean.CleanCmd       `cmd:"clean"    name:"clean"    hidden:"true" help:"The clean command is used to clean related processes."`
	YasdbCmd yasdb.YasdbCmd       `cmd:"yasdb"    name:"yasdb"    hidden:"true" help:"The yasdb command is used to manage yasshandb information."`
	Report   report.ReportCmd     `cmd:"report"   name:"report"   hidden:"true" help:"The report command is used to generate new reports from collection result."`
}
