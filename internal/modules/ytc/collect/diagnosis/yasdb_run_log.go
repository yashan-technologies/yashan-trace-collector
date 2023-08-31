package diagnosis

import (
	"fmt"
	"path"
	"strings"
	"time"

	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
)

func (b *DiagCollecter) collectYasdbRunLog() (err error) {
	yasdbRunLogItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_RUNLOG}
	defer b.fillResult(&yasdbRunLogItem)

	log := log.Module.M(datadef.DIAG_YASDB_RUNLOG)
	log.Debug("start to collect yasdb run.log")
	runLogPath, runLogFile := path.Join(b.YasdbData, LOG_DIR_NAME, YASDB_RUN_LOG), fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_RUN_LOG)
	if !b.notConnectDB {
		if runLogPath, err = GetYasdbRunLogPath(b.CollectParam); err != nil {
			log.Error(err)
			yasdbRunLogItem.Error = err.Error()
			yasdbRunLogItem.Description = datadef.GenGetDatabaseParameterDesc(string(yasdb.PM_RUN_LOG_FILE_PATH))
			return
		}
	}
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)
	// get run log files
	runLogFiles, err := b.getLogFiles(log, runLogPath, YASDB_RUN_LOG)
	if err != nil {
		log.Error(err)
		yasdbRunLogItem.Error = err.Error()
		yasdbRunLogItem.Description = datadef.GenNoPermissionDesc(runLogPath)
		return
	}
	// write run log to dest
	if err = b.collectRunLog(log, runLogFiles, path.Join(destPath, runLogFile), b.StartTime, b.EndTime); err != nil {
		log.Error(err)
		yasdbRunLogItem.Error = err.Error()
		yasdbRunLogItem.Description = datadef.GenDefaultDesc()
		return
	}
	yasdbRunLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME, runLogFile))
	return
}

func (b *DiagCollecter) collectRunLog(log yaslog.YasLog, srcs []string, dest string, start, end time.Time) (err error) {
	timeParseFunc := func(date time.Time, line string) (t time.Time, err error) {
		fields := strings.Split(line, stringutil.STR_BLANK_SPACE)
		if len(fields) < 2 {
			err = fmt.Errorf("invalid line: %s, skip", line)
			return
		}
		timeStr := fmt.Sprintf("%s %s", fields[0], fields[1])
		return time.ParseInLocation(timedef.TIME_FORMAT_WITH_MICROSECOND, timeStr, time.Local)
	}
	for _, f := range srcs {
		logEndTime := time.Now()
		if path.Base(f) != fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_RUN_LOG) {
			fileds := strings.Split(strings.TrimSuffix(path.Base(f), ".log"), stringutil.STR_HYPHEN)
			if len(fileds) < 2 {
				log.Errorf("failed to get log end time from %s, skip", f)
				continue
			}
			if logEndTime, err = time.ParseInLocation(timedef.TIME_FORMAT_IN_FILE, fileds[1], time.Local); err != nil {
				log.Errorf("failed to parse log end time from %s", fileds[1])
				continue
			}
		}
		if logEndTime.Before(b.StartTime) {
			// no need to write into dest
			log.Debugf("skip run log file: %s", f)
			continue
		}
		if err = b.collectLog(log, f, dest, time.Now(), timeParseFunc); err != nil {
			return
		}
	}
	return
}
