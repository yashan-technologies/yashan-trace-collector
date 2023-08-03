package resultgenner

import (
	"errors"
	"fmt"
	"os"
	"path"

	"ytc/defs/bashdef"
	"ytc/defs/runtimedef"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/stringutil"
	"ytc/utils/userutil"

	"git.yasdb.com/go/yasutil/execer"
	"git.yasdb.com/go/yasutil/fs"
)

const (
	_PACKAGE_NAME_FORMATTER = "ytc-%d"
	_REPORT_NAME_FORMATTER  = "report-%d.%s"
	_DATA_NAME_FORMATTER    = "data-%d.json"

	_DIR_DATA = "data"
)

type BaseResultGenner struct {
	NodeDatas    map[string]interface{} // key: nodeid, value: node data
	CollectTypes map[string]struct{}
	OutputDir    string
	ReportType   string
	Timestamp    int64
	Genner       Genner
}

func (g *BaseResultGenner) GenResult() (string, error) {
	if err := g.mkdirs(); err != nil {
		return stringutil.STR_EMPTY, err
	}
	if err := g.Genner.GenData(g.NodeDatas, g.genDataPath()); err != nil {
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

func (g *BaseResultGenner) genNodeDir(nodeID string) string {
	return path.Join(g.genPackageDir(), nodeID)
}

func (g *BaseResultGenner) genDataPath() string {
	name := fmt.Sprintf(_DATA_NAME_FORMATTER, g.Timestamp)
	return path.Join(g.genPackageDir(), _DIR_DATA, name)
}

func (g *BaseResultGenner) mkdirs() error {
	if !fs.IsDirExist(g.OutputDir) {
		if err := fs.Mkdir(g.OutputDir); err != nil {
			return err
		}
		if userutil.IsCurrentUserRoot() {
			owner := runtimedef.GetExecuteableOwner()
			_ = os.Chown(g.OutputDir, owner.Uid, owner.Uid)
		}
	}
	if err := fs.Mkdir(path.Dir(g.genDataPath())); err != nil {
		return err
	}
	for nodeID := range g.NodeDatas {
		nodeDir := g.genNodeDir(nodeID)
		for collectType := range g.CollectTypes {
			if err := fs.Mkdir(path.Join(nodeDir, collectType)); err != nil {
				return err
			}
		}
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
	executer := execer.NewExecer(log.Logger)
	ret, _, stderr := executer.Exec("bash", "-c", command)
	if ret != 0 {
		return errors.New(stderr)
	}
	return nil
}
