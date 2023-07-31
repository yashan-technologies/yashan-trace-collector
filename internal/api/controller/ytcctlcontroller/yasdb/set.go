package yasdb

type setCmd struct {
	DisableInteraction bool `name:"disable-interaction" short:"d" help:"Disable interaction edit mode."`
}

// [Interface Func]
func (c setCmd) Run() error {
	return nil
}
