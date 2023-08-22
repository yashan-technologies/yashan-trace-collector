package baseinforeporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"

	"github.com/jedib0t/go-pretty/v6/table"
)

// validate interface
var _ commons.Reporter = (*HostFirewallReporter)(nil)

type HostFirewallReporter struct{}

func NewHostFirewallReporterReporter() HostFirewallReporter {
	return HostFirewallReporter{}
}

// [Interface Func]
func (r HostFirewallReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report host firewall status
	isFirewallStatusActive, err := commons.ParseBool(item.Name, item.Details, "parse firewall status")
	if err != nil {
		return
	}
	writer := r.genReportContentWriter(isFirewallStatusActive)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r HostFirewallReporter) genReportContentWriter(isFirewallStatusActive bool) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"防火墙状态"})
	message := "已关闭"
	if isFirewallStatusActive {
		message = "已开启"
	}
	tw.AppendRow(table.Row{message})
	return tw
}
