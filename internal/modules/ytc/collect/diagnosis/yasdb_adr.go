package diagnosis

import (
	"fmt"
	"path"
	"time"

	"ytc/defs/errdef"
	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"

	"git.yasdb.com/go/yasutil/fs"
)

func (b *DiagCollecter) collectYasdbADR() (err error) {
	yasdbADRItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_ADR}
	defer b.fillResult(&yasdbADRItem)

	log := log.Module.M(datadef.DIAG_YASDB_ADR)
	// default adr path
	adrPath := path.Join(b.YasdbData, DIAG_DIR_NAME)
	if !b.notConnectDB {
		if adrPath, err = GetAdrPath(b.CollectParam); err != nil {
			log.Error(err)
			yasdbADRItem.Error = err.Error()
			yasdbADRItem.Description = datadef.GenGetDatabaseParameterDesc(string(yasdb.PM_DIAGNOSTIC_DEST))
			return
		}
	}
	if !fs.IsDirExist(adrPath) {
		err = &errdef.ErrFileNotFound{Fname: adrPath}
		log.Error(err)
		yasdbADRItem.Error = err.Error()
		yasdbADRItem.Description = datadef.GenNoPermissionDesc(adrPath)
		return
	}
	// package adr to dest
	destPath := path.Join(_packageDir, DIAG_DIR_NAME)
	destFile := fmt.Sprintf("yasdb-diag-%s.tar.gz", time.Now().Format(timedef.TIME_FORMAT_IN_FILE))
	if err = fs.TarDir(adrPath, path.Join(destPath, destFile)); err != nil {
		log.Error(err)
		yasdbADRItem.Error = err.Error()
		yasdbADRItem.Description = datadef.GenDefaultDesc()
		return
	}
	yasdbADRItem.Details = fmt.Sprintf("./%s", path.Join(DIAG_DIR_NAME, destFile))
	return
}
