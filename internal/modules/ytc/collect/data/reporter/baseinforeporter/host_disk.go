package baseinforeporter

import (
	"encoding/json"
	"fmt"

	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"

	"git.yasdb.com/go/yaserr"
	"git.yasdb.com/go/yasutil/size"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shirou/gopsutil/disk"
)

// validate interface
var _ commons.Reporter = (*HostDiskReporter)(nil)

type HostDiskReporter struct{}

func NewHostDiskReporter() HostDiskReporter {
	return HostDiskReporter{}
}

// [Interface Func]
func (r HostDiskReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report cpu base info
	usages, err := r.parseDiskUsage(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse disk usages")
		return
	}
	writer := r.genReportContentWriter(usages)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r HostDiskReporter) parseDiskUsage(item datadef.YTCItem) (usages []baseinfo.DiskUsage, err error) {
	usages, ok := item.Details.([]baseinfo.DiskUsage)
	if !ok {
		tmp, ok := item.Details.([]map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: item.Name,
				Targets: []interface{}{
					[]disk.UsageStat{},
					[]map[string]interface{}{},
				},
				Current: item.Details,
			}
			err = yaserr.Wrapf(err, "convert disk info interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &usages); err != nil {
			err = yaserr.Wrapf(err, "unmarshal disk info")
			return
		}
	}
	return
}

func (r HostDiskReporter) genReportContentWriter(usages []baseinfo.DiskUsage) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{
		"磁盘设备",
		"文件系统类型",
		"磁盘大小",
		"已使用",
		"可用",
		"使用率",
		"挂载路径",
		"挂载选项",
	})
	for _, u := range usages {
		tw.AppendRow(table.Row{
			u.Device,
			u.Fstype,
			size.GenHumanReadableSize(float64(u.Total), 2),
			size.GenHumanReadableSize(float64(u.Used), 2),
			size.GenHumanReadableSize(float64(u.Free), 2),
			fmt.Sprintf("%.2f%%", u.UsedPercent),
			u.Path,
			u.MountOptions,
		})
		tw.AppendSeparator()
	}
	return tw
}
