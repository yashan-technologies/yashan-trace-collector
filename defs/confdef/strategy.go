package confdef

import (
	"path"
	"ytc/defs/errdef"
	"ytc/defs/runtimedef"

	"git.yasdb.com/go/yasutil/fs"
	"github.com/BurntSushi/toml"
)

var _strategyConf Strategy

type Collect struct {
	Range  string `toml:"range"`
	Output string `toml:"output"`
}

type Report struct {
	Type   string `toml:"type"`
	Output string `toml:"output"`
}

type Strategy struct {
	Collect Collect `toml:"collect"`
	Report  Report  `toml:"report"`
}

func GetStrategyConf() Strategy {
	return _strategyConf
}

func initStrategyConf(strategyConf string) error {
	if !fs.IsFileExist(strategyConf) {
		return &errdef.ErrFileNotFound{Fname: strategyConf}
	}
	if _, err := toml.DecodeFile(strategyConf, &_strategyConf); err != nil {
		return &errdef.ErrFileParseFailed{Fname: strategyConf, Err: err}
	}
	if !path.IsAbs(_strategyConf.Collect.Output) {
		_strategyConf.Collect.Output = path.Join(runtimedef.GetYTCHome(), _strategyConf.Collect.Output)
	}
	if !path.IsAbs(_strategyConf.Report.Output) {
		_strategyConf.Report.Output = path.Join(runtimedef.GetYTCHome(), _strategyConf.Report.Output)
	}
	return nil
}
