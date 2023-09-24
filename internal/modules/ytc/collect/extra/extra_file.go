package extra

import (
	"os"
	"path"
	"strings"

	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"

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
		extraFile.Description = datadef.GenDefaultDesc()
		return
	}
	destPartentDir := path.Join(_packageDir, EXTRA_DIR_NAME)
	excludeMap := b.genExcludeMap()
	for dir, realPath := range dirs {
		dest := path.Join(destPartentDir, dir)
		if err = ytccollectcommons.CopyDir(log, realPath, dest, excludeMap); err != nil {
			log.Error(err)
			log.Errorf("failed to copy dir %s to %s, err: %v", realPath, dest, err)
			extraFile.Error = err.Error()
			extraFile.Error = datadef.GenDefaultDesc()
			return
		}
	}
	for file, realPath := range files {
		if _, ok := excludeMap[realPath]; ok {
			log.Infof("skip exclude file: %s", realPath)
			continue
		}
		dest := path.Join(destPartentDir, file)
		if err = fs.CopyFile(realPath, dest); err != nil {
			log.Errorf("failed to copy file %s to %s, err: %v", realPath, dest, err)
			extraFile.Error = err.Error()
			extraFile.Error = datadef.GenDefaultDesc()
			return
		}
	}
	extraFile.Details = b.genDetails(dirs, files)
	return
}

func (b *ExtraCollecter) genDetails(dirs map[string]string, files map[string]string) (res map[string]string) {
	res = make(map[string]string)
	for k, v := range dirs {
		res[b.GenPackageRelativePath(path.Join(EXTRA_DIR_NAME, k))] = v
	}
	for k, v := range files {
		res[b.GenPackageRelativePath(path.Join(EXTRA_DIR_NAME, k))] = v
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

func (b *ExtraCollecter) transferPath(path string) (res string) {
	res = strings.ReplaceAll(path, stringutil.STR_FORWARD_SLASH, stringutil.STR_UNDER_SCORE)
	return
}
