package runtimedef

import (
	"os"
	"path"
	"path/filepath"

	"ytc/utils/stringutil"
)

const (
	_ENV_YTC_HOME       = "YTC_HOME"
	_ENV_YTC_DEBUG_MODE = "YTC_DEBUG_MODE"
)

const (
	_DIR_NAME_LOG     = "log"
	_DIR_NAME_STATIC  = "static"
	_DIR_NAME_SCRIPTS = "scripts"
)

var _ytcHome string

func GetYTCHome() string {
	return _ytcHome
}

func GetLogPath() string {
	return path.Join(_ytcHome, _DIR_NAME_LOG)
}

func GetStaticPath() string {
	return path.Join(_ytcHome, _DIR_NAME_STATIC)
}

func GetScriptsPath() string {
	return path.Join(_ytcHome, _DIR_NAME_SCRIPTS)
}

func setYTCHome(v string) {
	_ytcHome = v
}

func isDebugMode() bool {
	return !stringutil.IsEmpty(os.Getenv(_ENV_YTC_DEBUG_MODE))
}

func getYTCHomeEnv() string {
	return os.Getenv(_ENV_YTC_HOME)
}

// genYTCHomeFromEnv generates ${YTC_HOME} from env, using YTC_HOME env as YTCHome in debug mode.
func genYTCHomeFromEnv() (ytcHome string, err error) {
	ytcHomeEnv := getYTCHomeEnv()
	if isDebugMode() && !stringutil.IsEmpty(ytcHomeEnv) {
		ytcHomeEnv, err = filepath.Abs(ytcHomeEnv)
		if err != nil {
			return
		}
		ytcHome = ytcHomeEnv
		return
	}
	return
}

// genYTCHomeFromRelativePath generates ${YTC_HOME} from relative path to the executable bin.
// executable bin locates at ${YTC_HOME}/bin/${executable}
func genYTCHomeFromRelativePath() (ytcHome string, err error) {
	executeable, err := getExecutable()
	if err != nil {
		return
	}
	ytcHome, err = filepath.Abs(path.Dir(path.Dir(executeable)))
	return
}

func initYTCHome() (err error) {
	ytcHome, err := genYTCHomeFromEnv()
	if err != nil {
		return
	}
	if !stringutil.IsEmpty(ytcHome) {
		setYTCHome(ytcHome)
		return
	}
	ytcHome, err = genYTCHomeFromRelativePath()
	if err != nil {
		return
	}
	setYTCHome(ytcHome)
	return
}
