package diagnosis

import (
	"os"
	"path"
	"regexp"
	"strings"

	"ytc/defs/confdef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yasutil/fs"
)

const (
	CORE_DIRECT           = "core direct"
	CORE_REDIRECT_ABRT    = "core redirect abrt"
	CORE_REDIRECT_SYSTEMD = "core redirect systemd"
)

const (
	CORE_FILE_REGEXP_str = `%[a-zA-Z]`
	ALL_REGEXP_STR       = `.*`
)

var coreFormatMap = map[string]string{
	`%p`: `\d*`,
	`%u`: `\d*`,
	`%g`: `\d*`,
	`%s`: `\d*`,
	`%t`: `\d*`,
	`%h`: `.*`,
	`%e`: `.*`,
	`%E`: `.*`,
	`%a`: `.*`,
	`%c`: `.*`,
}

func (b *DiagCollecter) yasdbCoreDumpFile() (err error) {
	yasdbCoreDumpItem := datadef.YTCItem{Name: datadef.DIAG_YASDB_COREDUMP}
	defer b.fillResult(&yasdbCoreDumpItem)

	log := log.Module.M(datadef.DIAG_YASDB_COREDUMP)
	originCoreDumpPath, coreDumpType, err := GetCoreDumpPath()
	coreDumpPath := originCoreDumpPath
	if err != nil {
		log.Errorf("failed to get coredump file path, err: %v", err)
		yasdbCoreDumpItem.Error = err.Error()
		yasdbCoreDumpItem.Description = datadef.GenGetCoreDumpPathDesc()
		return
	}
	if coreDumpType == CORE_DIRECT {
		if !path.IsAbs(originCoreDumpPath) {
			coreDumpPath = path.Join(b.YasdbHome, ytccollectcommons.BIN, originCoreDumpPath)
		}
		coreDumpPath = path.Dir(coreDumpPath)
	}
	log.Infof("coredump file path is: %s", coreDumpPath)
	re, err := b.getCoreDumpRegexp(originCoreDumpPath, coreDumpType)
	if err != nil {
		log.Errorf("failed to get coredump file name regexp, err: %v", err)
		yasdbCoreDumpItem.Error = err.Error()
		yasdbCoreDumpItem.Description = datadef.GenDefaultDesc()
		return
	}
	log.Infof("coredump file regexp is: %s", re.String())
	files, err := os.ReadDir(coreDumpPath)
	if err != nil {
		log.Errorf("failed to open dir: %s, err: %v", coreDumpPath, err)
		yasdbCoreDumpItem.Error = err.Error()
		yasdbCoreDumpItem.Description = datadef.GenReadCoreDumpPathDesc(coreDumpPath)
		return
	}
	for _, file := range files {
		if !file.Type().IsRegular() || !re.MatchString(file.Name()) {
			continue
		}
		if err := fileutil.CheckAccess(path.Join(coreDumpPath, file.Name())); err != nil {
			log.Errorf("does not have permission to %s, err: %v, skip collect the core file", file.Name(), err)
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
			log.Infof("the modify time of %s is %s, skip", file.Name(), createAt)
			continue
		}
		src, dest := path.Join(coreDumpPath, file.Name()), path.Join(_packageDir, ytccollectcommons.YASDB_DIR_NAME, CORE_DUMP_DIR_NAME, file.Name())
		if err = fs.CopyFile(src, dest); err != nil {
			log.Errorf("failed to copy file %s to %s", src, dest, err)
			yasdbCoreDumpItem.Error = err.Error()
			yasdbCoreDumpItem.Description = datadef.GenDefaultDesc()
			return
		}
	}
	yasdbCoreDumpItem.Details = b.GenPackageRelativePath(path.Join(ytccollectcommons.YASDB_DIR_NAME, CORE_DUMP_DIR_NAME))
	return
}

func (d *DiagCollecter) getCoreDumpRegexp(coreDumpPath string, coreDumpType string) (*regexp.Regexp, error) {
	coreFileKey := confdef.GetStrategyConf().Collect.CoreFileKey
	if !stringutil.IsEmpty(coreFileKey) {
		return regexp.Compile(coreFileKey)
	}
	key := CORE_FILE_KEY
	if coreDumpType == CORE_DIRECT {
		key = d.replaceCoreDumpPattern(path.Base(coreDumpPath))
	}
	return regexp.Compile(key)
}

func (d *DiagCollecter) getCoreDumpRealPath(originCoreDumpPath string, coreDumpType string) string {
	coreDumpPath := confdef.GetStrategyConf().Collect.CoreDumpPath
	if !stringutil.IsEmpty(coreDumpPath) {
		log.Module.Infof("coredump path in config is: %s", coreDumpPath)
		return coreDumpPath
	}
	coreDumpPath = originCoreDumpPath
	if coreDumpType == CORE_DIRECT {
		if !path.IsAbs(originCoreDumpPath) {
			coreDumpPath = path.Join(d.YasdbHome, ytccollectcommons.BIN, originCoreDumpPath)
		}
		coreDumpPath = path.Dir(coreDumpPath)
	}
	return coreDumpPath
}

func (d *DiagCollecter) replaceCoreDumpPattern(corePattern string) string {
	for k, v := range coreFormatMap {
		corePattern = strings.ReplaceAll(corePattern, k, v)
	}
	// replace other sub string does not contain in the coreFormatMap like "%M" with ".*"
	re := regexp.MustCompile(CORE_FILE_REGEXP_str)
	return re.ReplaceAllString(corePattern, ALL_REGEXP_STR)
}
