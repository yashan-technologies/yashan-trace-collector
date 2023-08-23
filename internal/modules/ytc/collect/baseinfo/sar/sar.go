package sar

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	"ytc/defs/runtimedef"
	"ytc/defs/timedef"
	"ytc/utils/execerutil"
	"ytc/utils/osutil"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
)

const (
	diskStatPath = "/proc/diskstats"
)

type Sar struct {
	log    yaslog.YasLog
	parser SarParser
}

func NewSar(yaslog yaslog.YasLog) *Sar {
	return &Sar{
		log:    yaslog,
		parser: getParser(yaslog),
	}
}

func getParser(yaslog yaslog.YasLog) SarParser {
	os := runtimedef.GetOSRelease()
	switch os.Id {
	case osutil.KYLIN_ID:
		return NewKylinParser(yaslog)
	case osutil.UBUNTU_ID:
		return NewUbuntuParser(yaslog)
	case osutil.CENTOS_ID:
		return NewCentosParser(yaslog)
	default:
		return NewBaseParser(yaslog)
	}
}

func (s *Sar) GetSarDir() string {
	return s.parser.GetSarDir()
}

func (s *Sar) Collect(t collecttypedef.WorkloadType, args ...string) (collecttypedef.WorkloadOutput, error) {
	res := make(collecttypedef.WorkloadOutput)
	execer := execerutil.NewExecer(s.log)
	realArgs := []string{bashdef.CMD_SAR}
	realArgs = append(realArgs, args...)
	cmd := fmt.Sprintf(" %s", strings.Join(realArgs, stringutil.STR_BLANK_SPACE))
	ret, stdout, stderr := execer.EnvExec(_envs, bashdef.CMD_BASH, "-c", cmd)
	if ret != 0 {
		err := errors.New(stderr)
		return res, err
	}
	parseFunc, checkTitleFunc := s.parser.GetParserFunc(t)
	res = s.parseSarOutput(stdout, parseFunc, checkTitleFunc)
	if t == collecttypedef.WT_DISK { // transfer Dev name
		var err error
		res, err = s.transferDiskOutput(res)
		if err != nil {
			err := errors.New(stderr)
			s.log.Error(err)
			return res, err
		}
	}
	return res, nil
}

// devNum is mainNum-subNum
func (s *Sar) genDevNumToDevNameMap() (map[string]string, error) {
	m := make(map[string]string)
	execer := execerutil.NewExecer(s.log)
	ret, stdout, stderr := execer.Exec(bashdef.CMD_CAT, diskStatPath)
	if ret != 0 {
		err := fmt.Errorf("failed to transfer dev number to dev name, err: %s", stderr)
		s.log.Error(err)
		return m, err
	}
	stdout = strings.ReplaceAll(stdout, "\n\n", stringutil.STR_NEWLINE)
	lines := strings.Split(stdout, stringutil.STR_NEWLINE)
	for _, line := range lines {
		if stringutil.IsEmpty(line) {
			continue
		}
		line := stringutil.RemoveExtraSpaces(strings.TrimSpace(line))
		values := strings.Split(line, stringutil.STR_BLANK_SPACE)
		if len(values) < 3 { // main number, sub number, dev name
			s.log.Warnf("invalid line: %s, skip it", line)
			continue
		}
		m[fmt.Sprintf("dev%s-%s", values[0], values[1])] = values[2]
	}
	return m, nil
}

func (s *Sar) transferDiskOutput(output collecttypedef.WorkloadOutput) (collecttypedef.WorkloadOutput, error) {
	m, err := s.genDevNumToDevNameMap()
	if err != nil {
		s.log.Error(err)
		return output, err
	}
	for timestamp, sarItem := range output {
		newItem := make(collecttypedef.WorkloadItem)
		for devNum, data := range sarItem {
			disk, ok := data.(DiskIO) // transfer interface to disk
			if !ok {
				s.log.Errorf("invaild data type: %v, skip it", reflect.TypeOf(data))
				continue
			}
			devName, ok := m[devNum] // get dev name
			if !ok {
				s.log.Errorf("can not find dev name by dev number: %s, skip it", devNum)
				continue
			}
			disk.Dev = devName
			newItem[devName] = disk
		}
		output[timestamp] = newItem
	}
	return output, nil
}

// example: Linux 3.10.0-1160.el7.x86_64 (mg_4)     08/10/2023      _x86_64_        (4 CPU)
func (s *Sar) getDateFromHeadLine(line string) string {
	currentData := time.Now().Format(timedef.TIME_FORMAT_DATE)
	line = stringutil.RemoveExtraSpaces(strings.TrimSpace(line))
	values := strings.Split(line, stringutil.STR_BLANK_SPACE)
	if len(values) < 4 { // third item is the data
		s.log.Error("invalid head line: %s, could not get data from the line", line)
		return currentData
	}
	arr := strings.Split(values[3], "/")
	// the third item is the data
	if len(arr) != 3 {
		s.log.Errorf("invalid data str: %s, could not get real data from the data str", values[3])
		return currentData
	}
	return fmt.Sprintf("%s-%s-%s", arr[2], arr[0], arr[1])
}

// getSarTimestamp get timestamp from sar time output like '11:52:42 AM' and data like '2023:08:08'
func (s *Sar) getSarTimestamp(data string, timeStr string, dayPeriod string) (int64, error) {
	t, err := time.ParseInLocation(timedef.TIME_FORMAT, fmt.Sprintf("%s %s", data, timeStr), time.Local)
	if err != nil {
		s.log.Error(err)
		return 0, err
	}
	if dayPeriod == timedef.DAY_PERIOD_PM {
		t = t.Add(12 * time.Hour)
	}
	return t.UTC().Unix(), nil
}

func (s *Sar) parseSarOutput(output string, parseFunc SarParseFunc, checkTitleFunc SarCheckTitleFunc) collecttypedef.WorkloadOutput {
	date := time.Now().Format(timedef.TIME_FORMAT_DATE)
	res := make(collecttypedef.WorkloadOutput)
	output = strings.ReplaceAll(output, "\n\n", stringutil.STR_NEWLINE)
	lines := strings.Split(output, stringutil.STR_NEWLINE)
	for i := 0; i < len(lines); i++ {
		line := stringutil.RemoveExtraSpaces(strings.TrimSpace(lines[i]))
		if strings.HasPrefix(line, LINUX_PREFIX) { // get data from head line
			date = s.getDateFromHeadLine(line)
			continue
		}
		// ignore the empty line, Average line and title line
		if stringutil.IsEmpty(line) || strings.HasPrefix(line, AVERAGE_PREFIX) || checkTitleFunc(line) {
			continue
		}
		values := strings.Split(line, stringutil.STR_BLANK_SPACE)
		if len(values) < 2 { // not enough data, skip
			s.log.Warnf("not enough data, skip line: %s", line)
			continue
		}
		// get time
		timestamp, err := s.getSarTimestamp(date, values[0], values[1])
		if err != nil {
			s.log.Error(err)
			continue
		}
		// get data map
		m, ok := res[timestamp]
		if !ok {
			m = make(collecttypedef.WorkloadItem)
		}
		res[timestamp] = parseFunc(m, values)
	}
	return res
}
