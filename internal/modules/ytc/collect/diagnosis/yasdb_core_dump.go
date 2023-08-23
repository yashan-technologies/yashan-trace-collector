package diagnosis

import (
	"fmt"
	"os"
	"path"
	"strings"

	"ytc/defs/confdef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yasutil/fs"
)

func (b *DiagCollecter) yasdbCoredumpFile() (err error) {
	yasdbCoreDumpItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_COREDUMP}
	defer b.fillResult(&yasdbCoreDumpItem)

	log := log.Module.M(datadef.DIAG_YASDB_COREDUMP)
	coreDumpPath, err := GetCoredumpPath()
	if err != nil {
		log.Error(err)
		yasdbCoreDumpItem.Error = err.Error()
		return
	}
	if !path.IsAbs(coreDumpPath) {
		coreDumpPath = path.Join(b.YasdbHome, "bin", coreDumpPath)
	}
	log.Infof("core dump file path is: %s", coreDumpPath)
	coreFileKey := confdef.GetStrategyConf().Collect.CoreFileKey
	if stringutil.IsEmpty(coreFileKey) {
		coreFileKey = CORE_FILE_KEY
	}
	files, err := os.ReadDir(coreDumpPath)
	if err != nil {
		log.Error(err)
		yasdbCoreDumpItem.Error = err.Error()
		return
	}
	for _, file := range files {
		if !file.Type().IsRegular() || !strings.Contains(file.Name(), coreFileKey) {
			continue
		}
		info, e := file.Info()
		if e != nil {
			err = e
			log.Error(err)
			yasdbCoreDumpItem.Error = err.Error()
			return
		}
		createAt := info.ModTime()
		if createAt.Before(b.StartTime) || createAt.After(b.EndTime) {
			continue
		}
		if err = fs.CopyFile(path.Join(coreDumpPath, file.Name()), path.Join(_packageDir, DIAG_DIR_NAME, CORE_DUMP_DIR_NAME, file.Name())); err != nil {
			log.Error(err)
			yasdbCoreDumpItem.Error = err.Error()
			return
		}
	}
	yasdbCoreDumpItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, CORE_DUMP_DIR_NAME))
	return
}
