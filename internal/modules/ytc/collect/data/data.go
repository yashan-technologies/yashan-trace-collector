package data

import (
	"fmt"
	"time"

	"ytc/defs/collecttypedef"
	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	report "ytc/internal/modules/ytc/collect/data/reporter"
	"ytc/internal/modules/ytc/collect/resultgenner"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/log"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
)

// validate interface
var _ resultgenner.Genner = (*YTCReport)(nil)

type YTCReport struct {
	CollectBeginTime time.Time                     `json:"collectBeginTime"`
	CollectEndtime   time.Time                     `json:"collectEndTime"`
	CollectParam     *collecttypedef.CollectParam  `json:"collectParam"`
	Modules          map[string]*datadef.YTCModule `json:"modules"`
	genner           resultgenner.BaseGenner
}

func NewYTCReport(param *collecttypedef.CollectParam) *YTCReport {
	return &YTCReport{
		CollectParam: param,
		Modules:      make(map[string]*datadef.YTCModule),
		genner:       resultgenner.BaseGenner{},
	}
}

// [Interface Func]
func (r *YTCReport) GenData(data interface{}, fname string) error {
	return r.genner.GenData(data, fname)
}

// [Interface Func]
func (r *YTCReport) GenReport() (content reporter.ReportContent, err error) {
	logger := log.Module.M("generate report")
	moduleNum := 0
	for _, moduleName := range _moduleOrder {
		module, ok := r.Modules[moduleName]
		if !ok {
			logger.Infof("module: %s unfound, pass", moduleName)
			continue
		}
		moduleNum++

		moduleTitlePrefix := fmt.Sprintf("%d", moduleNum)
		moduleContent := reporter.GenReportContentByTitle(fmt.Sprintf("%s %s", moduleTitlePrefix, collecttypedef.CollectTypeChineseName[moduleName]), reporter.FONT_SIZE_H1)
		content.Txt += moduleContent.Txt
		content.Markdown += moduleContent.Markdown
		content.HTML += moduleContent.HTML

		itemNum := 0
		items := module.Items()
		for _, itemName := range _itemOrder[moduleName] {
			item, ok := items[itemName]
			if !ok {
				logger.Infof("item: %s unfound, pass", itemName)
				continue
			}
			reporter, ok := report.REPORTERS[itemName]
			if !ok {
				err = fmt.Errorf("reporter of %s unfound", itemName)
				err = yaserr.Wrapf(err, "get reporter")
				return
			}
			itemNum++

			itemTitlePrefix := moduleTitlePrefix + stringutil.STR_DOT + fmt.Sprintf("%d", itemNum)
			itemContent, e := reporter.Report(*item, itemTitlePrefix)
			if e != nil {
				err = yaserr.Wrapf(e, "generete report of %s", itemName)
				return
			}
			content.Txt += itemContent.Txt + stringutil.STR_NEWLINE
			content.Markdown += itemContent.Markdown + stringutil.STR_NEWLINE
			content.HTML += itemContent.HTML + stringutil.STR_NEWLINE
		}
	}
	content.HTML += reporter.HTML_CSS
	return
}

func (r *YTCReport) GenResult(outputDir string, types map[string]struct{}) (string, error) {
	for _, m := range r.Modules {
		m.FillJSONItems()
	}
	genner := resultgenner.BaseResultGenner{
		Datas:        r.Modules,
		CollectTypes: types,
		OutputDir:    outputDir,
		Timestamp:    r.CollectBeginTime.Format(timedef.TIME_FORMAT_IN_FILE),
		Genner:       r,
	}
	return genner.GenResult()
}

func (r *YTCReport) GetPackageDir() string {
	genner := resultgenner.BaseResultGenner{
		OutputDir: r.CollectParam.Output,
		Timestamp: r.CollectBeginTime.Format(timedef.TIME_FORMAT_IN_FILE),
	}
	return genner.GetPackageDir()
}
