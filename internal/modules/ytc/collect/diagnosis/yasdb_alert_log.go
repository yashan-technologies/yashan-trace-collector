package diagnosis

import (
	"fmt"
	"path"
	"strings"
	"time"

	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/stringutil"
)

func (b *DiagCollecter) collectYasdbAlertLog() (err error) {
	yasdbAlertLogItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_ALERTLOG}
	defer b.fillResult(&yasdbAlertLogItem)

	log := log.Module.M(datadef.DIAG_YASDB_ALERTLOG)
	logPath := path.Join(b.YasdbData, LOG_DIR_NAME)
	alertLogPath, alertLogFile := path.Join(logPath, YASDB_ALERT_LOG), fmt.Sprintf(LOG_FILE_SUFFIX, YASDB_ALERT_LOG)
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)
	// get alert log
	timeParseFunc := func(date time.Time, line string) (t time.Time, e error) {
		fields := strings.Split(line, stringutil.STR_BAR)
		return time.ParseInLocation(timedef.TIME_FORMAT_WITH_MICROSECOND, fields[0], time.Local)
	}
	srcFile, destFile := path.Join(alertLogPath, alertLogFile), path.Join(destPath, alertLogFile)
	if err = b.collectLog(log, srcFile, destFile, time.Now(), timeParseFunc); err != nil {
		log.Error(err)
		yasdbAlertLogItem.Error = err.Error()
		yasdbAlertLogItem.Description = datadef.GenDefaultDesc()
		return
	}
	yasdbAlertLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME, alertLogFile))
	return
}
