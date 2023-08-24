package extra

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/fileutil"
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
	for dir, realPath := range dirs {
		dest := path.Join(destPartentDir, dir)
<<<<<<< Updated upstream
		if err = b.copyDir(log, realPath, dest, excludeMap); err != nil {
			log.Error(err)
=======
		if err = b.copyDir(log, realPath, dest, excludeMap); err != nil && b.isFileNameTooLongErr(err) {
			log.Errorf("failed to copy dir %s to %s, err: %v", realPath, dest, err)
>>>>>>> Stashed changes
			extraFile.Error = err.Error()
			return
		}
	}
	for file, realPath := range files {
		if _, ok := excludeMap[realPath]; ok {
			log.Infof("skip exclude file: %s", realPath)
			continue
		}
		dest := path.Join(destPartentDir, file)
<<<<<<< Updated upstream
		if err = fs.CopyFile(realPath, dest); err != nil {
			log.Error(err)
=======
		if err = fs.CopyFile(realPath, dest); err != nil && b.isFileNameTooLongErr(err) {
			log.Errorf("failed to copy file %s to %s, err: %v", realPath, dest, err)
>>>>>>> Stashed changes
			extraFile.Error = err.Error()
			return
		}
	}
	extraFile.Details = b.genDetails(dirs, files)
	return
}

func (b *ExtraCollecter) genDetails(dirs map[string]string, files map[string]string) (res map[string]string) {
	res = make(map[string]string)
	for k, v := range dirs {
		res[fmt.Sprintf("./%s", path.Join(EXTRA_DIR_NAME, k))] = v
	}
	for k, v := range files {
		res[fmt.Sprintf("./%s", path.Join(EXTRA_DIR_NAME, k))] = v
	}
	return
}

func (b *ExtraCollecter) filterInclude() (dirs map[string]string, files map[string]string, err error) {
	pathToInfoMap := make(map[string]os.FileInfo)
	dirs = make(map[string]string)
	files = make(map[string]string)
	for _, path := range b.Include {
		var info os.FileInfo
		info, err = os.Stat(path)
		if err != nil {
			return
		}
		pathToInfoMap = b.mergePath(path, info, pathToInfoMap)
	}
	for k, v := range pathToInfoMap {
		if v.IsDir() {
			dirName := v.Name()
			realPath, ok := dirs[dirName]
			if !ok {
				dirs[dirName] = k
				continue
			}
			// compare depth
			if fileutil.ComparePathDepth(k, realPath) > 0 {
				delete(dirs, dirName)
				dirs[b.transferPath(realPath)] = realPath
				dirs[dirName] = k
			} else {
				dirs[b.transferPath(k)] = k
			}
		} else {
			fileName := v.Name()
			realPath, ok := files[fileName]
			if !ok {
				files[fileName] = k
				continue
			}
			// compare depth
			if fileutil.ComparePathDepth(k, realPath) > 0 {
				delete(files, fileName)
				files[b.transferPath(realPath)] = realPath
				files[fileName] = k
			} else {
				files[b.transferPath(k)] = k
			}
		}
	}
	return
}

// mergePath merge input Include, if include contains /tmp and /tmp/test.go, it will return only /tmp
func (b *ExtraCollecter) mergePath(filePath string, info os.FileInfo, m map[string]os.FileInfo) map[string]os.FileInfo {
	m[filePath] = info
	for k, v := range m {
		if filePath == k {
			continue
		}
		dir := filePath
		if info.Mode().IsRegular() {
			dir = path.Dir(filePath)
		}
		if fileutil.IsAncestorDir(k, dir) && v.IsDir() {
			delete(m, filePath)
			continue
		}
		if fileutil.IsAncestorDir(dir, k) && info.IsDir() {
			delete(m, k)
			continue
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
			return filepath.SkipDir
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

func (b *ExtraCollecter) isFileNameTooLongErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "file name too long")
}
