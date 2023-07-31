package confdef

import (
	"path"
	"ytc/defs/errdef"
	"ytc/defs/runtimedef"

	"git.yasdb.com/go/yasutil/fs"
	"github.com/BurntSushi/toml"
)

var _ytcConf Ytc

type Ytc struct {
	StrategyPath string `toml:"strategy_path"`
	LogLevel     string `toml:"log_level"`
}

func GetYTCConf() Ytc {
	return _ytcConf
}

func initYTCConf(ytcConf string) error {
	if !path.IsAbs(ytcConf) {
		ytcConf = path.Join(runtimedef.GetYTCHome(), ytcConf)
	}
	if !fs.IsFileExist(ytcConf) {
		return &errdef.ErrFileNotFound{Fname: ytcConf}
	}
	if _, err := toml.DecodeFile(ytcConf, &_ytcConf); err != nil {
		return &errdef.ErrFileParseFailed{Fname: ytcConf, Err: err}
	}
	if !path.IsAbs(_ytcConf.StrategyPath) {
		_ytcConf.StrategyPath = path.Join(runtimedef.GetYTCHome(), _ytcConf.StrategyPath)
	}
	return nil
}
