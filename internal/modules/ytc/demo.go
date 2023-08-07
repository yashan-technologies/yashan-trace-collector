package ytc

import (
	"fmt"
	"strings"
	"time"
	"ytc/defs/bashdef"
	"ytc/internal/modules/ytc/resultgenner"
	"ytc/utils/stringutil"
)

type DemoBaseResult struct {
	Uname string `json:"uname"`
}

type NodeDemoResult struct {
	Base DemoBaseResult `json:"base"`
}

type DemoResults struct {
	Results map[string]*NodeDemoResult // key: node id, value: result
	genner  resultgenner.BaseGenner
}

// [Interface Func]
func (d *DemoResults) GenData(data interface{}, fname string) error {
	return d.genner.GenData(data, fname)
}

// [Interface Func]
func (d *DemoResults) GenReport() []byte {
	var content string
	for node, result := range d.Results {
		content += fmt.Sprintf("node: %s, uname: %s%s", node, result.Base.Uname, stringutil.STR_NEWLINE)
	}
	content = strings.TrimSuffix(content, stringutil.STR_NEWLINE)
	return []byte(content)
}

func (d *DemoResults) GenResult(outputDir, reportType string, types map[string]struct{}) (string, error) {
	genner := resultgenner.BaseResultGenner{
		NodeDatas:    make(map[string]interface{}),
		CollectTypes: types,
		OutputDir:    outputDir,
		ReportType:   reportType,
		Timestamp:    time.Now().Format(bashdef.TIME_FORMATTER),
		Genner:       d,
	}
	for k, v := range d.Results {
		genner.NodeDatas[k] = v
	}
	return genner.GenResult()
}
