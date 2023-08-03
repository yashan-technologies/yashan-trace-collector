package daemon

type DaemonCmd struct {
	Start   startCmd   `cmd:"start"   name:"start"   help:"Start yashan trace collector daemon."`
	Stop    stopCmd    `cmd:"stop"    name:"stop"    help:"Stop yashan trace collector daemon."`
	Restart restartCmd `cmd:"restart" name:"restart" help:"Restart yashan trace collector daemon."`
	Status  statusCmd  `cmd:"status"  name:"status"  help:"Show yashan trace collector daemon status."`
	Reload  reloadCmd  `cmd:"reload"  name:"reload"  help:"Reload yashan trace collector daemon."`
}

// [Interface Func]
func (c DaemonCmd) Run() error {
	return nil
}
