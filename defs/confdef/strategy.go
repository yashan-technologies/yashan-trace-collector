package confdef

import (
	"path"
	"regexp"
	"strings"
	"time"

	"ytc/defs/errdef"
	"ytc/defs/runtimedef"
	"ytc/utils/stringutil"
	"ytc/utils/timeutil"

	"git.yasdb.com/go/yasutil/fs"
	"github.com/BurntSushi/toml"
)

var _strategyConf Strategy

type Collect struct {
	Range              string `toml:"range"`
	Output             string `toml:"output"`
	MaxDuration        string `toml:"max_duration"`
	MinDuration        string `toml:"min_duration"`
	ScrapeInterval     int    `toml:"scrape_interval"`
	ScrapeTimes        int    `toml:"scrape_times"`
	ProcessNumberLimit int    `toml:"process_number_limit"`
	SarDir             string `toml:"sar_dir"`
	CoreFileKey        string `toml:"core_file_key"`
	CoreDumpPath       string `toml:"core_dump_path"`
	NetworkIODiscard   string `toml:"network_io_discard"`
	AWRTimeout         string `toml:"awr_timeout"`
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

func (c Collect) GetMaxDuration() (time.Duration, error) {
	if len(c.MaxDuration) == 0 {
		return time.Hour * 24, nil
	}
	maxDuration, err := timeutil.GetDuration(c.MaxDuration)
	if err != nil {
		return 0, err
	}
	return maxDuration, err
}

func (c Collect) GetMinDuration() (time.Duration, error) {
	if len(c.MinDuration) == 0 {
		return time.Minute * 1, nil
	}
	minDuration, err := timeutil.GetDuration(c.MinDuration)
	if err != nil {
		return 0, err
	}
	return minDuration, err
}

func (c Collect) GetMinAndMaxDur() (min time.Duration, max time.Duration, err error) {
	min, err = c.GetMinDuration()
	if err != nil {
		return
	}
	max, err = c.GetMaxDuration()
	if err != nil {
		return
	}
	return
}

func (c Collect) GetRange() (r time.Duration) {
	r, err := timeutil.GetDuration(c.Range)
	if err != nil {
		return time.Hour * 24
	}
	return
}

func (c Collect) GetNetworkIODiscard() []string {
	return strings.Split(c.NetworkIODiscard, stringutil.STR_COMMA)
}

func IsDiscardNetwork(name string) bool {
	discards := strings.Split(_strategyConf.Collect.NetworkIODiscard, stringutil.STR_COMMA)
	for _, discard := range discards {
		re, err := regexp.Compile(discard)
		if err != nil {
			continue
		}
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func (c Collect) GetAWRTimeout() (t time.Duration) {
	var err error
	t, err = timeutil.GetDuration(c.AWRTimeout)
	if err != nil {
		t = time.Minute * 24
		return
	}
	return
}
