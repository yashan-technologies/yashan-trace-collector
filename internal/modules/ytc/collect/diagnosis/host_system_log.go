package diagnosis

import (
	"fmt"
	"path"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/userutil"
)

func (b *DiagCollecter) collectHostSystemLog() (err error) {
	hostSystemLogItem := datadef.YTCItem{
		Name:     datadef.DIAG_HOST_SYSTEMLOG,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostSystemLogItem)

	log := log.Module.M(datadef.DIAG_HOST_SYSTEMLOG)
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)
	if userutil.IsCurrentUserRoot() {
		// message.log
		destMessageLogFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_MESSAGES_LOG))
		if err = b.collectHostLog(log, SYSTEM_LOG_MESSAGES, destMessageLogFile, SYSTEM_MESSAGES_LOG); err != nil {
			log.Error(err)
			hostSystemLogItem.Children[SYSTEM_MESSAGES_LOG] = datadef.YTCItem{Error: err.Error(), Description: datadef.GenDefaultDesc()}
		} else {
			logPath := b.GenPackageRelativePath(path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_MESSAGES_LOG)))
			hostSystemLogItem.Children[SYSTEM_MESSAGES_LOG] = datadef.YTCItem{Details: logPath}
		}
		// syslog.log
		destSysLogFile := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_SYS_LOG))
		if err = b.collectHostLog(log, SYSTEM_LOG_SYSLOG, destSysLogFile, SYSTEM_SYS_LOG); err != nil {
			log.Error(err)
			hostSystemLogItem.Children[SYSTEM_SYS_LOG] = datadef.YTCItem{Error: err.Error(), Description: datadef.GenDefaultDesc()}
		} else {
			logPath := b.GenPackageRelativePath(path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_SYS_LOG)))
			hostSystemLogItem.Children[SYSTEM_SYS_LOG] = datadef.YTCItem{Details: logPath}
		}
	} else {
		message := "has no permission to collect system log"
		description := datadef.GenNoPermissionSyslogDesc()
		hostSystemLogItem.Children[SYSTEM_MESSAGES_LOG] = datadef.YTCItem{
			Error:       message,
			Description: description,
		}
		hostSystemLogItem.Children[SYSTEM_SYS_LOG] = datadef.YTCItem{
			Error:       message,
			Description: description,
		}
	}
	return
}
