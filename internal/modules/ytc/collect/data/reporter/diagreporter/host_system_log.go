package diagreporter

import (
	"fmt"
	"strings"

	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/diagnosis"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
)

// validate interface
var _ commons.Reporter = (*HostSystemLogReporter)(nil)

type HostSystemLogReporter struct{}

func NewHostSystemLogReporter() HostSystemLogReporter {
	return HostSystemLogReporter{}
}

// [Interface Func]
func (r HostSystemLogReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, diagnosis.DiagChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2
	txt := reporter.GenTxtTitle(title)
	markdown := reporter.GenMarkdownTitle(title, fontSize)
	html := reporter.GenHTMLTitle(title, fontSize)

	messageLogItem, sysLogItem, err := r.validateSystemLogItem(item)
	if err != nil {
		err = yaserr.Wrapf(err, "validate host system log")
		return
	}

	messageLogContent, err := r.genMessageLogContent(messageLogItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate host message log content")
		return
	}

	sysLogContent, err := r.genSysLogContent(sysLogItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate host sys log content")
		return
	}

	content.Txt = strings.Join([]string{txt, messageLogContent.Txt, sysLogContent.Txt}, stringutil.STR_NEWLINE)
	content.Markdown = strings.Join([]string{markdown, messageLogContent.Markdown, sysLogContent.Markdown}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{html, messageLogContent.HTML, sysLogContent.HTML}, stringutil.STR_NEWLINE)
	return
}

func (r HostSystemLogReporter) validateSystemLogItem(item datadef.YTCItem) (messageLogItem, sysLogItem datadef.YTCItem, err error) {
	if len(item.Children) == 0 {
		err = fmt.Errorf("invalid data, children of %v unfound", item)
		return
	}
	messageLogItem, ok := item.Children[diagnosis.SYSTEM_MESSAGES_LOG]
	if !ok {
		err = fmt.Errorf("invalid data, %s unfound in %v", diagnosis.SYSTEM_MESSAGES_LOG, item.Children)
		return
	}
	sysLogItem, ok = item.Children[diagnosis.SYSTEM_SYS_LOG]
	if !ok {
		err = fmt.Errorf("invalid data, %s unfound in %v", diagnosis.SYSTEM_SYS_LOG, item.Children)
		return
	}
	return
}

func (r HostSystemLogReporter) parseMessageLogItem(messageLogItem datadef.YTCItem) (messageLog string, err error) {
	return commons.ParseString(diagnosis.SYSTEM_MESSAGES_LOG, messageLogItem.Details, "parse host message log")
}

func (r HostSystemLogReporter) parseSysLogItem(sysLogItem datadef.YTCItem) (sysLog string, err error) {
	return commons.ParseString(diagnosis.SYSTEM_SYS_LOG, sysLogItem.Details, "parse host sys log")
}

func (r HostSystemLogReporter) genMessageLogContent(messageLogItem datadef.YTCItem, titlePrefix string) (messageLogItemContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.1 %s", titlePrefix, diagnosis.SYSTEM_MESSAGES_LOG)
	fontSize := reporter.FONT_SIZE_H3
	if len(messageLogItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(messageLogItem.Error, messageLogItem.Description)
		messageLogItemContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		messageLog, e := r.parseMessageLogItem(messageLogItem)
		if e != nil {
			err = yaserr.Wrapf(e, "parse host message log")
			return
		}
		tw := commons.GenPathWriter(messageLog)
		messageLogItemContent = reporter.GenReportContentByWriterAndTitle(tw, title, fontSize)
	}
	return
}

func (r HostSystemLogReporter) genSysLogContent(sysLogItem datadef.YTCItem, titlePrefix string) (sysLogContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.2 %s", titlePrefix, diagnosis.SYSTEM_SYS_LOG)
	fontSize := reporter.FONT_SIZE_H3
	if len(sysLogItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(sysLogItem.Error, sysLogItem.Description)
		sysLogContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		sysLog, e := r.parseSysLogItem(sysLogItem)
		if e != nil {
			err = yaserr.Wrapf(e, "parse host sys log")
			return
		}
		tw := commons.GenPathWriter(sysLog)
		sysLogContent = reporter.GenReportContentByWriterAndTitle(tw, title, fontSize)
	}
	return
}
