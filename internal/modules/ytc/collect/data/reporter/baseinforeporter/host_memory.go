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
	"github.com/shirou/gopsutil/mem"
)

// validate interface
var _ commons.Reporter = (*HostMemoryReporter)(nil)

type HostMemoryReporter struct{}

func NewHostMemoryReporter() HostMemoryReporter {
	return HostMemoryReporter{}
}

// [Interface Func]
func (r HostMemoryReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2

	// report error
	if len(item.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(item.Error, item.Description)
		content = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
		return
	}

	// report memory base info
	memory, err := r.parseMemory(item)
	if err != nil {
		err = yaserr.Wrapf(err, "parse memory")
		return
	}
	writer := r.genReportContentWriter(memory)
	content = reporter.GenReportContentByWriterAndTitle(writer, title, fontSize)
	return
}

func (r HostMemoryReporter) parseMemory(item datadef.YTCItem) (memory *mem.VirtualMemoryStat, err error) {
	memory, ok := item.Details.(*mem.VirtualMemoryStat)
	if !ok {
		tmp, ok := item.Details.(map[string]interface{})
		if !ok {
			err = &commons.ErrInterfaceTypeNotMatch{
				Key: item.Name,
				Targets: []interface{}{
					&mem.VirtualMemoryStat{},
					map[string]interface{}{},
				},
				Current: item.Details,
			}
			err = yaserr.Wrapf(err, "parse memory info interface")
			return
		}
		data, _ := json.Marshal(tmp)
		if err = json.Unmarshal(data, &memory); err != nil {
			err = yaserr.Wrapf(err, "unmarshal memory info")
			return
		}
	}
	return
}

func (r HostMemoryReporter) genReportContentWriter(memory *mem.VirtualMemoryStat) reporter.Writer {
	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{
		"",
		"内存大小",
		"已使用",
		"空闲",
		"共享内存",
		"缓冲/缓存",
		"可用",
	})

	tw.AppendRow(table.Row{
		"系统内存",
		size.GenHumanReadableSize(float64(memory.Total), 2),
		size.GenHumanReadableSize(float64(memory.Used), 2),
		size.GenHumanReadableSize(float64(memory.Free), 2),
		size.GenHumanReadableSize(float64(memory.Shared), 2),
		size.GenHumanReadableSize(float64(memory.Buffers+memory.Cached), 2),
		size.GenHumanReadableSize(float64(memory.Available), 2),
	})
	tw.AppendSeparator()

	swapUsed := memory.SwapTotal - memory.SwapFree + memory.SwapCached
	tw.AppendRow(table.Row{
		"交换分区",
		size.GenHumanReadableSize(float64(memory.SwapTotal), 2),
		size.GenHumanReadableSize(float64(swapUsed), 2),
		size.GenHumanReadableSize(float64(memory.SwapFree), 2),
		"/",
		size.GenHumanReadableSize(float64(memory.SwapCached), 2),
	})
	return tw
}
