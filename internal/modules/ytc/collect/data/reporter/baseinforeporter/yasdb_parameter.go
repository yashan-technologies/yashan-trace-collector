package baseinforeporter

import (
	"encoding/json"
	"fmt"
	"strings"

	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
)

// validate interface
var _ commons.Reporter = (*YashanDBParameterReporter)(nil)

type YashanDBParameterReporter struct{}

func NewYashanDBParameterReporter() YashanDBParameterReporter {
	return YashanDBParameterReporter{}
}

// [Interface Func]
func (r YashanDBParameterReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2
	txt := reporter.GenTxtTitle(title)
	markdown := reporter.GenMarkdownTitle(title, fontSize)
	html := reporter.GenHTMLTitle(title, fontSize)

	yasdbIni, parameter, err := r.validateParameterItem(item)
	if err != nil {
		err = yaserr.Wrapf(err, "validate yasdb parameter")
		return
	}

	yasdbIniContent, err := r.genYasdbIniContent(yasdbIni, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate yasdb.ini content")
		return
	}

	parameterContent, err := r.genVParameterContent(parameter, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate v$parameter content")
		return
	}

	content.Txt = strings.Join([]string{txt, yasdbIniContent.Txt, parameterContent.Txt}, stringutil.STR_NEWLINE)
	content.Markdown = strings.Join([]string{markdown, yasdbIniContent.Markdown, parameterContent.Markdown}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{html, yasdbIniContent.HTML, parameterContent.HTML}, stringutil.STR_NEWLINE)
	return
}

func (r YashanDBParameterReporter) validateParameterItem(item datadef.YTCItem) (yasdbIni, parameter datadef.YTCItem, err error) {
	if len(item.Children) == 0 {
		err = fmt.Errorf("invalid data, children of %s unfound", item.Name)
		return
	}
	yasdbIni, ok := item.Children[baseinfo.KEY_YASDB_INI]
	if !ok {
		err = fmt.Errorf("invalid data, %s unfound in %v", baseinfo.KEY_YASDB_INI, item.Children)
		return
	}
	parameter, ok = item.Children[baseinfo.KEY_YASDB_PARAMETER]
	if !ok {
		err = fmt.Errorf("invalid data, %s unfound in %v", baseinfo.KEY_YASDB_PARAMETER, item.Children)
		return
	}
	return
}

func (r YashanDBParameterReporter) parseYasdbIni(yasdbIni datadef.YTCItem) (ymap map[string]string, err error) {
	ymap, ok := yasdbIni.Details.(map[string]string)
	if !ok {
		tmp, ok := yasdbIni.Details.(map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: baseinfo.KEY_YASDB_INI,
				Targets: []interface{}{
					map[string]string{},
					map[string]interface{}{},
				},
				Current: yasdbIni.Details,
			}
			err = yaserr.Wrapf(err, "parse yasdb.ini interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &ymap); err != nil {
			err = yaserr.Wrapf(err, "unmarshal yasdb.ini")
			return
		}
	}
	return
}

func (r YashanDBParameterReporter) parseParameter(parameter datadef.YTCItem) (parameters []*yasdb.VParameter, err error) {
	parameters, ok := parameter.Details.([]*yasdb.VParameter)
	if !ok {
		tmp, ok := parameter.Details.([]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: baseinfo.KEY_YASDB_PARAMETER,
				Targets: []interface{}{
					[]*yasdb.VParameter{},
					[]interface{}{},
				},
				Current: parameter.Details,
			}
			err = yaserr.Wrapf(err, "parse v$parameter interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &parameters); err != nil {
			err = yaserr.Wrapf(err, "unmarshal v$parameter")
			return
		}
	}
	return
}

func (r YashanDBParameterReporter) genYasdbIniContent(yasdbIni datadef.YTCItem, titlePrefix string) (yasdbIniContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.1 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_YASDB_INI])
	fontSize := reporter.FONT_SIZE_H3
	if len(yasdbIni.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(yasdbIni.Error, yasdbIni.Description)
		yasdbIniContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		ymap, e := r.parseYasdbIni(yasdbIni)
		if e != nil {
			err = yaserr.Wrapf(e, "parse yasdb.ini")
			return
		}
		tw := commons.ReporterWriter.NewTableWriter()
		tw.AppendHeader(table.Row{"参数名称", "参数值"})
		for key, val := range ymap {
			tw.AppendRow(table.Row{key, val})
			tw.AppendSeparator()
		}
		yasdbIniContent = reporter.GenReportContentByWriterAndTitle(tw, title, fontSize)
	}
	return
}

func (r YashanDBParameterReporter) genVParameterContent(parameter datadef.YTCItem, titlePrefix string) (parameterContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.2 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_YASDB_PARAMETER])
	fontSize := reporter.FONT_SIZE_H3
	if len(parameter.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(parameter.Error, parameter.Description)
		parameterContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		parameters, e := r.parseParameter(parameter)
		if e != nil {
			err = yaserr.Wrapf(e, "parse v$parameter")
			return
		}
		tw := commons.ReporterWriter.NewTableWriter()
		tw.AppendHeader(table.Row{"参数名称", "参数值"})
		for _, p := range parameters {
			tw.AppendRow(table.Row{p.Name, p.Value})
			tw.AppendSeparator()
		}
		parameterContent = reporter.GenReportContentByWriterAndTitle(tw, title, fontSize)
	}
	return
}
