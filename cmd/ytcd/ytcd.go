// This is the main package for ytcd.
// Ytcd means ytc daemon, which is used to do some scheduled collector works.
package main

import (
	"ytc/commons/flags"
	"ytc/defs/compiledef"
	"ytc/defs/confdef"
	"ytc/defs/runtimedef"

	"github.com/alecthomas/kong"
)

const (
	_APP_NAME        = "ytcd"
	_APP_DESCRIPTION = "Ytcd means ytc daemon, which is used to do some scheduled collector works."
)

func main() {
	var app App
	options := flags.NewAppOptions(_APP_NAME, _APP_DESCRIPTION, compiledef.GetAPPVersion())
	ctx := kong.Parse(&app, options...)
	if err := initApp(app); err != nil {
		ctx.FatalIfErrorf(err)
	}
	if err := ctx.Run(); err != nil {
		ctx.FatalIfErrorf(err)
	}
}

func initApp(app App) error {
	if err := runtimedef.InitRuntime(); err != nil {
		return err
	}
	if err := confdef.InitConf(app.Config); err != nil {
		return err
	}
	return nil
}
