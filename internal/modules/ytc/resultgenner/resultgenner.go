package resultgenner

import (
	"errors"
	"fmt"
	"os"
	"path"

	"ytc/defs/bashdef"
	"ytc/defs/runtimedef"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"
	"ytc/utils/userutil"

	"git.yasdb.com/go/yasutil/fs"
)

const (
	_PACKAGE_NAME_FORMATTER = "ytc-%s"
	_REPORT_NAME_FORMATTER  = "report-%s.%s"
	_DATA_NAME_FORMATTER    = "data-%s.json"

	_DIR_DATA   = "data"
	_DIR_BASE   = "base"
	_DIR_DIAG   = "diag"
	_DIR_PERF   = "perf"
	_DIR_LOG    = "log"
	_DIR_YASDB  = "yasdb"
	_DIR_SYSTEM = "system"
)

type BaseResultGenner struct {
	Datas        interface{}
	CollectTypes map[string]struct{}
	OutputDir    string
	ReportType   string
	Timestamp    string
	Genner       Genner
}

func (g *BaseResultGenner) GenResult() (string, error) {
	if err := g.Mkdirs(); err != nil {
		return stringutil.STR_EMPTY, err
	}
	if err := g.Genner.GenData(g.Datas, g.genDataPath()); err != nil {
		log.Module.Warnf("generate data failed: %s", err)
	}
	if err := g.writeReport(); err != nil {
		log.Module.Errorf("write report failed: %s", err)
		return stringutil.STR_EMPTY, err
	}
	if err := g.tarResult(); err != nil {
		log.Module.Errorf("tar result failed: %s", err)
		return stringutil.STR_EMPTY, err
	}
	return g.genPackageTarPath(), nil
}

func (g *BaseResultGenner) GetPackageDir() string {
	return g.genPackageDir()
}

func (g *BaseResultGenner) genPackageName() string {
	return fmt.Sprintf(_PACKAGE_NAME_FORMATTER, g.Timestamp)
}

func (g *BaseResultGenner) genPackageDir() string {
	return path.Join(g.OutputDir, g.genPackageName())
}

func (g *BaseResultGenner) genPackageTarName() string {
	return fmt.Sprint(g.genPackageName(), ".tar.gz")
}

func (g *BaseResultGenner) genPackageTarPath() string {
	return path.Join(g.OutputDir, g.genPackageTarName())
}

func (g *BaseResultGenner) genDataPath() string {
	name := fmt.Sprintf(_DATA_NAME_FORMATTER, g.Timestamp)
	return path.Join(g.genPackageDir(), _DIR_DATA, name)
}

func (g *BaseResultGenner) Mkdirs() error {
	if !fs.IsDirExist(g.OutputDir) {
		if err := fs.Mkdir(g.OutputDir); err != nil {
			return err
		}
		if userutil.IsCurrentUserRoot() {
			owner := runtimedef.GetExecuteableOwner()
			if owner.Uid != 0 || owner.Gid != 0 {
				_ = os.Chown(g.OutputDir, owner.Uid, owner.Uid)
			}
		}
	}
	if err := fs.Mkdir(path.Dir(g.genDataPath())); err != nil {
		return err
	}
	return nil
}

func (g *BaseResultGenner) genReportPath() string {
	name := fmt.Sprintf(_REPORT_NAME_FORMATTER, g.Timestamp, g.ReportType)
	return path.Join(g.genPackageDir(), name)
}

func (g *BaseResultGenner) writeReport() error {
	content := g.Genner.GenReport()
	return fileutil.WriteFile(g.genReportPath(), content)
}

func (g *BaseResultGenner) tarResult() error {
	command := fmt.Sprintf("cd %s;%s czvf %s %s;rm -rf %s", g.OutputDir, bashdef.CMD_TAR, g.genPackageTarName(), g.genPackageName(), g.genPackageName())
	executer := execerutil.NewExecer(log.Logger)
	ret, _, stderr := executer.Exec("bash", "-c", command)
	if ret != 0 {
		return errors.New(stderr)
	}
	return nil
}
