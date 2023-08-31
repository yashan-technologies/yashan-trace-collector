package diagnosis

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"ytc/defs/timedef"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"
	"ytc/utils/timeutil"

	"git.yasdb.com/go/yaslog"
)

func (b *DiagCollecter) collectHostLog(log yaslog.YasLog, src, dest string, prefix string) (err error) {
	hasSetDateext, err := b.hasSetDateext()
	if err != nil {
		return
	}
	if hasSetDateext {
		return b.collectHostLogWithSetDateext(log, src, dest, prefix)
	}
	return b.collectHostLogWithoutSetDateext(log, src, dest)
}

func (b *DiagCollecter) hostLogTimeParse(date time.Time, line string) (t time.Time, err error) {
	fields := strings.Split(line, stringutil.STR_BLANK_SPACE)
	if len(fields) < 3 {
		err = fmt.Errorf("invalid line: %s, skip", line)
		return
	}
	tmpTime, err := time.ParseInLocation(timedef.TIME_FORMAT_TIME, fields[2], time.Local)
	if err != nil {
		return
	}
	hour, min, sec := tmpTime.Hour(), tmpTime.Minute(), tmpTime.Second()
	day, err := strconv.Atoi(fields[1])
	if err != nil {
		return
	}
	mon, err := timeutil.GetMonth(fields[0])
	year := date.Year()
	if date.Month() < mon {
		year = year - 1
	}
	t = time.Date(year, mon, day, hour, min, sec, 0, time.Local)
	return
}

func (b *DiagCollecter) collectHostLogWithSetDateext(log yaslog.YasLog, src, dest string, prefix string) (err error) {
	var srcs []string
	srcs, err = b.getLogFiles(log, path.Dir(src), prefix)
	if err != nil {
		return
	}
	var logFiles []string // resort logFile so that the current log file is the last one, other file sorted by time is in the first
	for _, v := range srcs {
		if v == src {
			continue
		}
		logFiles = append(logFiles, v)
	}
	if len(srcs) != len(logFiles) {
		logFiles = append(logFiles, src)
	}
	for _, logFile := range logFiles {
		log.Debugf("try to collect %s", logFile)
		date := time.Now()
		if logFile != src {
			fileds := strings.Split(path.Base(logFile), stringutil.STR_HYPHEN)
			if len(fileds) < 2 {
				log.Errorf("failed to get log end time from %s, skip", logFile)
				continue
			}
			// get date from log file name
			date, err = time.ParseInLocation(timedef.TIME_FORMAT_DATE_IN_FILE, fileds[1], time.Local)
			if err != nil {
				log.Error("failed to get date from: %s, err: %s", logFile, err.Error())
				continue
			}
			// try to get log end time from last 3 line in log
			k := 3
			lastKLines, err := fileutil.Tail(logFile, k)
			if err != nil {
				log.Errorf("failed to read file %s last %d line, err: %s", logFile, k, err.Error())
			} else {
				for i := 0; i < len(lastKLines); i++ {
					if stringutil.IsEmpty(lastKLines[i]) {
						continue
					}
					var tmpData time.Time
					tmpData, err = b.hostLogTimeParse(date, stringutil.RemoveExtraSpaces(strings.TrimSpace(lastKLines[i])))
					if err != nil {
						log.Errorf("failed to parse time from line: %s, err: %s", lastKLines[i], err.Error())
						continue
					}
					date = tmpData
				}
			}
			log.Debugf("log file %s end date is %s", logFile, date)
			if date.Before(b.StartTime) {
				log.Infof("skip to collect log %s, log file end date: %s , collect start date %s", logFile, date.AddDate(0, 0, -1), b.StartTime)
				continue
			}
		}
		if err = b.collectLog(log, logFile, dest, date, b.hostLogTimeParse); err != nil {
			log.Errorf("failed to collect from: %s, err: %s", logFile, err.Error())
			continue
		}
		log.Debugf("succeed to collect %s", logFile)
	}
	return
}

func (b *DiagCollecter) collectHostLogWithoutSetDateext(log yaslog.YasLog, src, dest string) (err error) {
	// get log file last modify time
	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}
	srcModTime := srcInfo.ModTime()
	if srcModTime.Before(b.StartTime) {
		log.Infof("log %s last modify time is %s, skip", src, srcModTime)
		return
	}
	return b.reverseCollectLog(log, src, dest, srcModTime, b.hostLogTimeParse)
}

func (b *DiagCollecter) hasSetDateext() (res bool, err error) {
	config, err := os.Open(LOG_ROTATE_CONFIG)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(config)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "dateext") {
			res = true
			return
		}
	}
	return
}

func (b *DiagCollecter) getLogFiles(log yaslog.YasLog, logPath string, prefix string) (logFiles []string, err error) {
	entrys, err := os.ReadDir(logPath)
	if err != nil {
		log.Error(err)
		return
	}
	for _, entry := range entrys {
		if !entry.Type().IsRegular() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		logFiles = append(logFiles, path.Join(logPath, entry.Name()))
	}
	// sort with file name
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i] < logFiles[j]
	})
	return
}

// some log may not contain date info in the log file content, but in the log name
func (b *DiagCollecter) collectLog(log yaslog.YasLog, src, dest string, date time.Time, timeParseFunc logTimeParseFunc) (err error) {
	destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileutil.DEFAULT_FILE_MODE)
	if err != nil {
		return
	}
	defer destFile.Close()
	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer srcFile.Close()

	var t time.Time
	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		txt := scanner.Text()
		line := stringutil.RemoveExtraSpaces(strings.TrimSpace(txt))
		if stringutil.IsEmpty(line) {
			continue
		}
		if t, err = timeParseFunc(date, line); err != nil {
			log.Error("skip line: %s, err: %s", txt, err.Error())
			continue
		}
		if t.Before(b.StartTime) {
			continue
		}
		if t.After(b.EndTime) {
			break
		}
		_, err = destFile.WriteString(txt + stringutil.STR_NEWLINE)
		if err != nil {
			return
		}
	}
	log.Debugf("succeed to write log file %s to %s", src, dest)
	return
}

func (b *DiagCollecter) reverseCollectLog(log yaslog.YasLog, src, dest string, date time.Time, timeParseFunc logTimeParseFunc) (err error) {
	// open tmp file
	tmp := fmt.Sprintf("%s.temp", dest)
	tmpFile, err := os.OpenFile(tmp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileutil.DEFAULT_FILE_MODE)
	if err != nil {
		return
	}
	defer os.Remove(tmp)
	// open src file in reverse order
	reverseSrcFile, err := fileutil.NewReverseFile(src)
	if err != nil {
		return
	}
	defer reverseSrcFile.Close()
	for {
		line, e := reverseSrcFile.ReadLine()
		if e != nil {
			if e == io.EOF {
				// read to end
				break
			}
			err = e
			return
		}
		var t time.Time
		t, err = timeParseFunc(date, stringutil.RemoveExtraSpaces(strings.TrimSpace(line)))
		if err != nil {
			return
		}
		if t.After(b.EndTime) {
			continue
		}
		if t.Before(b.StartTime) {
			break
		}
		// write to tmp file
		if _, err = tmpFile.WriteString(line + stringutil.STR_NEWLINE); err != nil {
			return
		}
	}
	// reverse open tmp file
	reverseTmpFile, err := fileutil.NewReverseFile(tmp)
	if err != nil {
		return
	}
	defer reverseTmpFile.Close()
	// open dest file
	destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileutil.DEFAULT_FILE_MODE)
	if err != nil {
		return
	}
	defer destFile.Close()
	for {
		line, e := reverseTmpFile.ReadLine()
		if e != nil {
			if e == io.EOF {
				// read to end
				break
			}
			err = e
			return
		}
		// write to dest file
		if _, err = destFile.WriteString(line + stringutil.STR_NEWLINE); err != nil {
			return
		}
	}
	log.Debugf("succeed to write log file %s to %s", src, dest)
	return
}
