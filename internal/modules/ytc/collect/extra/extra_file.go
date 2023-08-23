package extra

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/fs"
)

func (b *ExtraCollecter) collectExtraFile() (err error) {
	extraFile := datadef.YTCItem{Name: datadef.EXTRA_FILE_COLLECT}
	defer b.fillResult(&extraFile)

	log := log.Module.M(datadef.EXTRA_FILE_COLLECT)
	dirs, files, err := b.filterInclude()
	if err != nil {
		log.Error(err)
		extraFile.Error = err.Error()
		return
	}
	destPartentDir := path.Join(_packageDir, EXTRA_DIR_NAME)
	excludeMap := b.genExcludeMap()
	for dir := range dirs {
		dest := path.Join(destPartentDir, b.transferPath(dir))
		if err = b.copyDir(log, dir, dest, excludeMap); err != nil {
			log.Error(err)
			extraFile.Error = err.Error()
			return
		}
	}
	for file := range files {
		if _, ok := excludeMap[file]; ok {
			log.Infof("skip exclude file: %s", file)
			continue
		}
		dest := path.Join(destPartentDir, b.transferPath(file))
		if err = fs.CopyFile(file, dest); err != nil {
			log.Error(err)
			extraFile.Error = err.Error()
			return
		}
	}
	extraFile.Details = fmt.Sprintf("./%s", EXTRA_DIR_NAME)
	return
}

func (b *ExtraCollecter) filterInclude() (dirs map[string]struct{}, files map[string]struct{}, err error) {
	m := make(map[string]os.FileInfo)
	dirs = make(map[string]struct{})
	files = make(map[string]struct{})
	for _, path := range b.Include {
		var info os.FileInfo
		info, err = os.Stat(path)
		if err != nil {
			return
		}
		m = b.mergePath(path, info, m)
	}
	for k, v := range m {
		if v.IsDir() {
			dirs[k] = struct{}{}
		}
		if v.Mode().IsRegular() {
			files[k] = struct{}{}
		}
	}
	return
}

// mergePath merge input Include, if include contains /tmp and /tmp/test.go, it will return only /tmp
func (b *ExtraCollecter) mergePath(path string, info os.FileInfo, m map[string]os.FileInfo) map[string]os.FileInfo {
	m[path] = info
	for k, v := range m {
		if path == k {
			continue
		}
		if strings.HasPrefix(path, k) && v.IsDir() {
			delete(m, path)
		}
		if strings.HasPrefix(k, path) && info.IsDir() {
			delete(m, k)
		}
	}
	return m
}

func (b *ExtraCollecter) genExcludeMap() (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, path := range b.Exclude {
		res[path] = struct{}{}
	}
	return
}

func (b *ExtraCollecter) copyDir(log yaslog.YasLog, src, dest string, excludeMap map[string]struct{}) (err error) {
	if strings.TrimSpace(src) == strings.TrimSpace(dest) {
		log.Infof("src path: %s is equal to dest path: %s, skip", src, dest)
		return
	}
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("failed to copy dir, err: %s", err.Error())
			return err
		}
		if _, ok := excludeMap[path]; ok {
			log.Infof("skip exclude path: %s", path)
			return nil
		}
		destNewPath := strings.Replace(path, src, dest, -1)
		if info.IsDir() {
			if err = os.MkdirAll(destNewPath, info.Mode()); err != nil {
				log.Errorf("failed to mkdir: %s, err: %s", destNewPath, err.Error())
				return err
			}
		} else {
			if err = fs.CopyFile(path, destNewPath); err != nil {
				// skip no permission file
				if os.IsPermission(err) {
					log.Infof("skip inaccessible path: %s", path)
					return nil
				}
				log.Errorf("failed to copy file %s to %s, err: %s", path, destNewPath, err.Error())
				return err
			}
		}
		return nil
	})
	return
}

func (b *ExtraCollecter) transferPath(path string) (res string) {
	res = strings.ReplaceAll(path, stringutil.STR_FORWARD_SLASH, stringutil.STR_UNDER_SCORE)
	return
}
