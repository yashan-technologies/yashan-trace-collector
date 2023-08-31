package diagnosis

import (
	"fmt"
	"path"

	"ytc/defs/bashdef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/fileutil"
)

func (b *DiagCollecter) collectHostKernelLog() (err error) {
	hostKernelLogItem := datadef.YTCItem{Name: datadef.DIAG_HOST_KERNELLOG}
	defer b.fillResult(&hostKernelLogItem)

	log := log.Module.M(datadef.DIAG_HOST_KERNELLOG)
	destPath := path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)
	// dmesg.log
	execer := execerutil.NewExecer(log)
	dmesgFile := fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_DMESG_LOG)
	dest := path.Join(destPath, fmt.Sprintf(LOG_FILE_SUFFIX, SYSTEM_DMESG_LOG))
	ret, stdout, stderr := execer.Exec(bashdef.CMD_BASH, "-c", bashdef.CMD_DMESG)
	if ret != 0 {
		err = fmt.Errorf("failed to get host dmesg log, err: %s", stderr)
		log.Error(err)
		hostKernelLogItem.Error = err.Error()
		hostKernelLogItem.Description = datadef.GenKylinDmesgDesc()
		return
	}
	// write to dest
	if err = fileutil.WriteFile(dest, []byte(stdout)); err != nil {
		log.Error(err)
		hostKernelLogItem.Error = err.Error()
		hostKernelLogItem.Description = datadef.GenDefaultDesc()
		return
	}
	hostKernelLogItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME, dmesgFile))
	return
}
