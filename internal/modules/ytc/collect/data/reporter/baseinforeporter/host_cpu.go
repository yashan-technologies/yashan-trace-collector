package baseinforeporter

import (
	"encoding/json"
	"fmt"

	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shirou/gopsutil/cpu"
)

// validate interface
var _ commons.Reporter = (*HostCPUReporter)(nil)

type HostCPUReporter struct{}

func NewHostCPUReporter() HostCPUReporter {
	return HostCPUReporter{}
}

// [Interface Func]
func (r HostCPUReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report cpu info
	cpuInfos, err := r.parseCPUInfos(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse cpu infos")
		return
	}
	writer := r.genReportContentWriter(cpuInfos)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r HostCPUReporter) parseCPUInfos(cpuInfoItem datadef.YTCItem) (cpuInfos []cpu.InfoStat, err error) {
	cpuInfos, ok := cpuInfoItem.Details.([]cpu.InfoStat)
	if !ok {
		tmp, ok := cpuInfoItem.Details.([]map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: cpuInfoItem.Name,
				Targets: []interface{}{
					[]cpu.InfoStat{},
					[]map[string]interface{}{},
				},
				Current: cpuInfoItem.Details,
			}
			err = yaserr.Wrapf(err, "convert cpu info interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &cpuInfos); err != nil {
			err = yaserr.Wrapf(err, "unmarshal cpu info")
			return
		}
	}
	return
}

func (r HostCPUReporter) genReportContentWriter(cpuInfos []cpu.InfoStat) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{"检查项", "检查结果"})

	tw.AppendRow(table.Row{"CPU型号", cpuInfos[0].ModelName})
	tw.AppendSeparator()

	var physicalCores, logicalCores int
	tmp := make(map[string]struct{})
	for _, c := range cpuInfos {
		tmp[c.PhysicalID] = struct{}{}
		logicalCores += int(c.Cores)
	}
	physicalCores = len(tmp)
	tw.AppendRow(table.Row{"CPU核心数量", fmt.Sprintf("物理CPU核心：%d，逻辑CPU核心：%d", physicalCores, logicalCores)})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"CPU主频", fmt.Sprintf("@%.2fGHz", cpuInfos[0].Mhz/1000)})
	tw.AppendSeparator()

	tw.AppendRow(table.Row{"厂商标识ID", cpuInfos[0].VendorID})

	return tw
}
