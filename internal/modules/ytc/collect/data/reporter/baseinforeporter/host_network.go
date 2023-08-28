package baseinforeporter

import (
	"encoding/json"
	"fmt"
	"strings"

	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
	"git.yasdb.com/go/yasutil/netcli"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shirou/gopsutil/net"
)

// validate interface
var _ commons.Reporter = (*HostNetworkReporter)(nil)

type HostNetworkReporter struct{}

func NewHostNetworkReporter() HostNetworkReporter {
	return HostNetworkReporter{}
}

// [Interface Func]
func (r HostNetworkReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report cpu base info
	networks, err := r.parseNetworkInfo(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse network info")
		return
	}
	content = r.genReportContentWriter(networks, title, fontSize)
	return
}

func (r HostNetworkReporter) parseNetworkInfo(item datadef.YTCItem) (networks []net.InterfaceStat, err error) {
	networks, ok := item.Details.([]net.InterfaceStat)
	if !ok {
		tmp, ok := item.Details.([]map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: item.Name,
				Targets: []interface{}{
					[]net.InterfaceStat{},
					[]map[string]interface{}{},
				},
				Current: item.Details,
			}
			err = yaserr.Wrapf(err, "parse netwotk info interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &networks); err != nil {
			err = yaserr.Wrapf(err, "unmarshal netwotk info")
			return
		}
	}
	return
}

func (r HostNetworkReporter) genReportContentWriter(networks []net.InterfaceStat, title string, fontSize reporter.FontSize) (content reporter.ReportContent) {
	titleContent := reporter.GenReportContentByTitle(title, fontSize)
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"网络接口", "IP地址", "MAC地址"})
	genTableRows := func(sep string) (rows []table.Row) {
		for _, n := range networks {
			var ipv4, ipv6 []string
			for _, addr := range n.Addrs {
				ip := addr.Addr
				if netcli.IsIPv6(ip) {
					ipv6 = append(ipv6, ip)
					continue
				}
				ipv4 = append(ipv4, ip)
			}
			var col []string
			if len(ipv4) > 0 {
				col = append(col, "IPv4: ")
				col = append(col, ipv4...)
			}
			if len(ipv6) > 0 {
				if len(col) > 0 {
					col = append(col, "")
				}
				col = append(col, "IPv6: ")
				col = append(col, ipv6...)
			}

			rows = append(rows, table.Row{n.Name, strings.Join(col, sep), n.HardwareAddr})
			tw.AppendSeparator()
		}
		return rows
	}

	// render txt
	for _, r := range genTableRows(stringutil.STR_NEWLINE) {
		tw.AppendRow(r)
		tw.AppendSeparator()
	}
	content.Txt = strings.Join([]string{titleContent.Txt, tw.Render()}, stringutil.STR_NEWLINE)
	tw.ResetRows()

	// render markdown and html
	tw.AppendRows(genTableRows(stringutil.STR_HTML_BR))
	content.Markdown = strings.Join([]string{titleContent.Markdown, tw.RenderMarkdown()}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{titleContent.HTML, tw.RenderHTML()}, stringutil.STR_NEWLINE)

	return
}
