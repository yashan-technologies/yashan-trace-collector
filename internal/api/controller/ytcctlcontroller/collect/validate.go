package collect

import (
	"errors"
	"os"
	"path"
	"strings"

	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	"ytc/defs/errdef"
	"ytc/defs/regexdef"
	"ytc/defs/runtimedef"
	"ytc/log"
	"ytc/utils/jsonutil"
	"ytc/utils/stringutil"
	"ytc/utils/timeutil"

	"github.com/google/uuid"
)

var (
	ErrOutPutNotPermission = errors.New("output permission denied")
)

func (c *CollectCmd) validate() error {
	strategyConf := confdef.GetStrategyConf()
	c.fillDefault(strategyConf)
	if err := c.validateType(); err != nil {
		return err
	}
	if err := c.validateRange(); err != nil {
		return err
	}
	if err := c.validateStartAndEnd(); err != nil {
		return err
	}
	if err := c.validateOutput(); err != nil {
		return err
	}
	if err := c.validateExtra(); err != nil {
		return err
	}
	return nil
}

func (c *CollectCmd) validateType() error {
	resMap := make(map[string]struct{})
	tMap := map[string]struct{}{
		collecttypedef.TYPE_BASE: {},
		collecttypedef.TYPE_DIAG: {},
		collecttypedef.TYPE_PREF: {},
	}
	types := strings.Split(c.Type, stringutil.STR_COMMA)
	for _, t := range types {
		if _, ok := tMap[t]; !ok {
			return errdef.NewErrFlagFormat(ytctl_collect, f_type)
		}
		resMap[t] = struct{}{}
	}
	return nil
}

func (c *CollectCmd) validateExtra() error {
	if err := c.validateExtraPath(c.Include); err != nil {
		return err
	}
	// no need to check exclude
	return nil
}

func (c *CollectCmd) validateExtraPath(value string) error {
	paths := c.getExtraPath(value)
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return err
		}
	}
	return nil
}

func (c *CollectCmd) validateRange() error {
	strategyConf := confdef.GetStrategyConf()
	log.Controller.Debugf("strategy: %s\n", jsonutil.ToJSONString(strategyConf))
	log.Controller.Debugf("cmd: %s", jsonutil.ToJSONString(c))
	if stringutil.IsEmpty(c.Range) {
		return nil
	}
	if !regexdef.RANGE_REGEX.MatchString(c.Range) {
		return errdef.NewErrFlagFormat(ytctl_collect, f_range)
	}
	minDuration, maxDuration, err := strategyConf.Collect.GetMinAndMaxDur()
	if err != nil {
		log.Controller.Errorf("get duration err: %s", err.Error())
		return err
	}
	log.Controller.Debugf("get min %s max %s", minDuration.String(), maxDuration.String())
	r := strategyConf.Collect.GetRange()
	if r > maxDuration {
		return errdef.NewGreaterMaxDur(strategyConf.Collect.MaxDuration)
	}
	if r < minDuration {
		return errdef.NewLessMinDur(strategyConf.Collect.MinDuration)
	}
	return nil
}

func (c *CollectCmd) validateStartAndEnd() error {
	strategyConf := confdef.GetStrategyConf()
	var startNotEmpty, endNotEmpty bool
	if !stringutil.IsEmpty(c.Start) {
		if !regexdef.TIME_REGEX.MatchString(c.Start) {
			return errdef.NewErrFlagFormat(ytctl_collect, f_start)
		}
		startNotEmpty = true
	}
	if !stringutil.IsEmpty(c.End) {
		if !regexdef.TIME_REGEX.MatchString(c.End) {
			return errdef.NewErrFlagFormat(ytctl_collect, f_end)
		}
		endNotEmpty = true
	}
	if startNotEmpty && endNotEmpty {
		minDuration, maxDuration, err := strategyConf.Collect.GetMinAndMaxDur()
		if err != nil {
			log.Controller.Errorf("get duration err: %s", err.Error())
			return err
		}
		start, err := timeutil.GetTimeDivBySepa(c.Start, stringutil.STR_HYPHEN)
		if err != nil {
			return err
		}
		end, err := timeutil.GetTimeDivBySepa(c.End, stringutil.STR_HYPHEN)
		if err != nil {
			return err
		}
		if end.Before(start) {
			return ErrEndLessStart
		}
		r := end.Sub(start)
		if r > maxDuration {
			return errdef.NewGreaterMaxDur(strategyConf.Collect.MaxDuration)
		}
		if r < minDuration {
			return errdef.NewLessMinDur(strategyConf.Collect.MaxDuration)
		}
	}
	return nil
}

func (c *CollectCmd) validateOutput() error {
	output := c.Output
	if !regexdef.PATH_REGEX.Match([]byte(output)) {
		return ErrPathFormat
	}
	if !path.IsAbs(output) {
		output = path.Join(runtimedef.GetYTCHome(), output)
	}
	_, err := os.Stat(output)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(output, 0766); err != nil {
			log.Controller.Errorf("create output err: %s", err.Error())
			return err
		}
	}
	tmpFile := uuid.NewString()[0:7]
	for {
		_, err := os.OpenFile(path.Join(output, tmpFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			log.Controller.Errorf("create tmp err: %s", err.Error())
			if os.IsPermission(err) {
				return errdef.NewErrPermissionDenied(output)
			}
			return err
		}
		_ = os.Remove(path.Join(output, tmpFile))
		return nil
	}
}

func (c *CollectCmd) fillDefault(stra confdef.Strategy) {
	if stringutil.IsEmpty(c.Output) {
		c.Output = confdef.GetStrategyConf().Collect.Output
	}
	if !path.IsAbs(c.Output) {
		c.Output = path.Join(runtimedef.GetYTCHome(), c.Output)
	}
}
