package baseinforeporter

import (
	"encoding/json"
	"fmt"
	"time"

	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shirou/gopsutil/host"
)

// validate interface
var _ commons.Reporter = (*HostOSInfoReporter)(nil)

type HostOSInfoReporter struct{}

func NewHostOSInfoReporter() HostOSInfoReporter {
	return HostOSInfoReporter{}
}

// [Interface Func]
func (r HostOSInfoReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report host info
	hostInfo, err := r.parseHostInfo(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse network info")
		return
	}
	writer := r.genReportContentWriter(hostInfo)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r HostOSInfoReporter) parseHostInfo(item datadef.YTCItem) (hostInfo *host.InfoStat, err error) {
	hostInfo, ok := item.Details.(*host.InfoStat)
	if !ok {
		tmp, ok := item.Details.(map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: item.Name,
				Targets: []interface{}{
					&host.InfoStat{},
					map[string]interface{}{},
				},
				Current: item.Details,
			}
			err = yaserr.Wrapf(err, "parse host info interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &hostInfo); err != nil {
			err = yaserr.Wrapf(err, "unmarshal host info")
			return
		}
	}
	return
}

func (r HostOSInfoReporter) genReportContentWriter(hostInfo *host.InfoStat) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"检查项", "检查结果"})

	tw.AppendRow(table.Row{"主机名称", hostInfo.Hostname})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"开机时间", time.Unix(int64(hostInfo.BootTime), 0).Format(timedef.TIME_FORMAT)})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"操作系统", hostInfo.OS})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"发行版本", fmt.Sprintf("%s %s (%s系列)", hostInfo.Platform, hostInfo.PlatformVersion, hostInfo.PlatformFamily)})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"内核版本", hostInfo.KernelVersion})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"内核架构", hostInfo.KernelArch})
	return tw
}
