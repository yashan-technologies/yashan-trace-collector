package yasdb

type YasdbCmd struct {
	Show showCmd `cmd:"show" name:"show" help:"Show current yashandb information."`
	Set  setCmd  `cmd:"set"  name:"set"  help:"Set yashandb information."`
}

// [Interface Func]
func (c YasdbCmd) Run() error {
	return nil
}
