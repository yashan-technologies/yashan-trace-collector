package baseinforeporter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/baseinfo/gopsutil"
	"ytc/internal/modules/ytc/collect/baseinfo/sar"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"
	"ytc/internal/modules/ytc/collect/resultgenner/reporter/htmldef"
	"ytc/utils/numutil"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
	"git.yasdb.com/go/yasutil/size"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	_graph_name_memory_usage = "内存使用率"

	_graph_history_memory_usage = "history_memory_usage"
	_graph_current_memory_usage = "current_memory_usage"
)

const (
	// keys
	_key_real_usage = "real_usage"

	// labels
	_label_real_usage = "真实使用率"
)

var (
	_yKeysMemory   = []string{_key_usage, _key_real_usage}
	_yLabelsMemory = []string{_label_usage, _label_real_usage}
)

// validate interface
var _ commons.Reporter = (*HostMemoryUsageReporter)(nil)

type HostMemoryUsageReporter struct{}

type sarMemoryUsage struct {
	timestamp int64
	sar.MemoryUsage
}

type gopsutilMemoryUsage struct {
	timestamp int64
	gopsutil.MemoryUsage
}

func NewHostMemoryUsageReporter() HostMemoryUsageReporter {
	return HostMemoryUsageReporter{}
}

// [Interface Func]
func (r HostMemoryUsageReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2
	txt := reporter.GenTxtTitle(title)
	markdown := reporter.GenMarkdownTitle(title, fontSize)
	html := reporter.GenHTMLTitle(title, fontSize)

	historyItem, currentItem, err := validateWorkLoadItem(item)
	if err != nil {
		err = yaserr.Wrapf(err, "validate memory usage item")
		return
	}

	historyItemContent, err := r.genHistoryContent(historyItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate memory usage history content")
		return
	}

	currentItemContent, err := r.genCurrentContent(currentItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate memory usage current content")
		return
	}

	content.Txt = strings.Join([]string{txt, historyItemContent.Txt, currentItemContent.Txt}, stringutil.STR_NEWLINE)
	content.Markdown = strings.Join([]string{markdown, historyItemContent.Markdown, currentItemContent.Markdown}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{html, historyItemContent.HTML, currentItemContent.HTML}, stringutil.STR_NEWLINE)
	content.Graph = strings.Join([]string{historyItemContent.Graph, currentItemContent.Graph}, stringutil.STR_NEWLINE)
	return
}

func (r HostMemoryUsageReporter) parseSarItem(item datadef.YTCItem) (output map[int64]map[string]sar.MemoryUsage, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal sar memory usage")
		return
	}
	output = make(map[int64]map[string]sar.MemoryUsage)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal sar memory usage")
		return
	}
	return
}

func (r HostMemoryUsageReporter) parseSarHistoryItem(historyItem datadef.YTCItem) (output map[int64]map[string]sar.MemoryUsage, err error) {
	output, err = r.parseSarItem(historyItem)
	if err != nil {
		err = yaserr.Wrapf(err, "history sar item")
	}
	return
}

func (r HostMemoryUsageReporter) parseSarCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]sar.MemoryUsage, err error) {
	output, err = r.parseSarItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current sar item")
	}
	return
}

func (r HostMemoryUsageReporter) parseGopsutilItem(item datadef.YTCItem) (output map[int64]map[string]gopsutil.MemoryUsage, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal gopsutil memory usage")
		return
	}
	output = make(map[int64]map[string]gopsutil.MemoryUsage)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal gopsutil memory usage")
		return
	}
	return
}

func (r HostMemoryUsageReporter) parseGopsutilCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]gopsutil.MemoryUsage, err error) {
	output, err = r.parseGopsutilItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current gopsutil item")
	}
	return
}

func (r HostMemoryUsageReporter) genSarReportContent(sarData map[int64]map[string]sar.MemoryUsage, graphName string) (content reporter.ReportContent) {
	tmp := make(map[string][]sarMemoryUsage)
	for time, val := range sarData {
		for k, v := range val {
			memoryUsage := sarMemoryUsage{
				timestamp:   time,
				MemoryUsage: v,
			}
			tmp[k] = append(tmp[k], memoryUsage)
		}
	}

	var keys []string
	for key := range tmp {
		keys = append(keys, key)
	}
	sort.StringSlice(keys).Sort()

	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{
		"时间",
		"总内存",
		"空闲内存",
		"已使用",
		"使用率",
		"缓冲(buffers)",
		"缓存(cache)",
		"可用",
		"已提交",
		"已提交的内存占总内存的百分比",
		"活跃内存的大小",
		"非活跃内存的大小",
		"已修改但尚未写入磁盘的脏页大小",
		"真实内存使用率",
	})
	for _, key := range keys {
		var rows []map[string]interface{}
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			totalKB := float64(p.KBmemUsed) / p.MemUsed * 100
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				size.GenHumanReadableSize(totalKB*1024, 2),
				size.GenHumanReadableSize(float64(p.KBMemFree*1024), 2),
				size.GenHumanReadableSize(float64(p.KBmemUsed*1024), 2),
				fmt.Sprintf("%.2f%%", p.MemUsed),
				size.GenHumanReadableSize(float64(p.KBBuffers*1024), 2),
				size.GenHumanReadableSize(float64(p.KBCached*1024), 2),
				size.GenHumanReadableSize(float64(p.KBAvail*1024), 2),
				size.GenHumanReadableSize(float64(p.KBCommit*1024), 2),
				fmt.Sprintf("%.2f%%", p.Commit),
				size.GenHumanReadableSize(float64(p.KBActive*1024), 2),
				size.GenHumanReadableSize(float64(p.KBInact*1024), 2),
				size.GenHumanReadableSize(float64(p.KBDirty*1024), 2),
				fmt.Sprintf("%.2f%%", p.RealMemUsed),
			})
			row := make(map[string]interface{})
			row[_key_time] = time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT)
			row[_key_usage] = numutil.TruncateFloat64(p.MemUsed, 2)
			row[_key_real_usage] = numutil.TruncateFloat64(p.RealMemUsed, 2)
			rows = append(rows, row)
		}
		c := reporter.GenReportContentByWriter(tw)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		content.HTML += reporter.GenHTMLTitle(_graph_name_memory_usage, reporter.FONT_SIZE_H4) + htmldef.GenGraphElement(graphName)
		content.Graph = htmldef.GenGraphData(graphName, rows, _key_time, _yKeysMemory, _yLabelsMemory)
		tw.ResetRows()
	}
	return
}

func (r HostMemoryUsageReporter) genGopsutilReportContent(gopsutilData map[int64]map[string]gopsutil.MemoryUsage, graphName string) (content reporter.ReportContent) {
	tmp := make(map[string][]gopsutilMemoryUsage)
	for time, val := range gopsutilData {
		for k, v := range val {
			memoryUsage := gopsutilMemoryUsage{
				timestamp:   time,
				MemoryUsage: v,
			}
			tmp[k] = append(tmp[k], memoryUsage)
		}
	}

	var keys []string
	for key := range tmp {
		keys = append(keys, key)
	}
	sort.StringSlice(keys).Sort()

	tw := commons.ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{
		"时间",
		"总内存",
		"空闲内存",
		"已使用",
		"使用率",
		"缓冲(buffers)",
		"缓存(cache)",
		"可用",
		"系统可保证的总内存使用量",
		"虚拟内存总大小",
		"虚拟内存已使用",
	})
	for _, key := range keys {
		var rows []map[string]interface{}
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				size.GenHumanReadableSize(float64(p.Total), 2),
				size.GenHumanReadableSize(float64(p.Free), 2),
				size.GenHumanReadableSize(float64(p.Used), 2),
				fmt.Sprintf("%.2f%%", p.UsedPercent),
				size.GenHumanReadableSize(float64(p.Buffers), 2),
				size.GenHumanReadableSize(float64(p.Cached), 2),
				size.GenHumanReadableSize(float64(p.Available), 2),
				size.GenHumanReadableSize(float64(p.CommitLimit), 2),
				size.GenHumanReadableSize(float64(p.VMallocTotal), 2),
				size.GenHumanReadableSize(float64(p.VMallocUsed), 2),
			})
			row := make(map[string]interface{})
			row[_key_time] = time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT)
			row[_key_usage] = numutil.TruncateFloat64(p.UsedPercent, 2)
			row[_key_usage] = numutil.TruncateFloat64(float64(p.Available/p.Total), 2)
			rows = append(rows, row)
		}
		c := reporter.GenReportContentByWriter(tw)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		content.HTML += reporter.GenHTMLTitle(_graph_name_memory_usage, reporter.FONT_SIZE_H4) + htmldef.GenGraphElement(graphName)
		content.Graph = htmldef.GenGraphData(graphName, rows, _key_time, _yKeysMemory, _yLabelsMemory)
		tw.ResetRows()
	}
	return
}

func (r HostMemoryUsageReporter) genHistoryContent(historyItem datadef.YTCItem, titlePrefix string) (historyItemContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.1 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_HISTORY])
	fontSize := reporter.FONT_SIZE_H3
	if len(historyItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(historyItem.Error, historyItem.Description)
		historyItemContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		historyItemContent = reporter.GenReportContentByTitle(title, fontSize)
		history, e := r.parseSarHistoryItem(historyItem)
		if e != nil {
			err = yaserr.Wrapf(e, "parse history memory usage")
			return
		}
		c := r.genSarReportContent(history, _graph_history_memory_usage)
		historyItemContent.Txt += c.Txt
		historyItemContent.Markdown += c.Markdown
		historyItemContent.HTML += c.HTML
		historyItemContent.Graph += c.Graph
	}
	return
}

func (r HostMemoryUsageReporter) genCurrentContent(currentItem datadef.YTCItem, titlePrefix string) (currentItemContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.2 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_CURRENT])
	fontSize := reporter.FONT_SIZE_H3
	if len(currentItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(currentItem.Error, currentItem.Description)
		currentItemContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		currentItemContent = reporter.GenReportContentByTitle(title, fontSize)
		if currentItem.DataType == datadef.DATATYPE_SAR {
			current, e := r.parseSarCurrentItem(currentItem)
			if e != nil {
				err = yaserr.Wrapf(e, "parse sar current memory usage")
				return
			}
			c := r.genSarReportContent(current, _graph_current_memory_usage)
			currentItemContent.Txt += c.Txt
			currentItemContent.Markdown += c.Markdown
			currentItemContent.HTML += c.HTML
			currentItemContent.Graph += c.Graph
		} else {
			current, e := r.parseGopsutilCurrentItem(currentItem)
			if e != nil {
				err = yaserr.Wrapf(e, "parse gopsutil current memory usage")
				return
			}
			c := r.genGopsutilReportContent(current, _graph_current_memory_usage)
			currentItemContent.Txt += c.Txt
			currentItemContent.Markdown += c.Markdown
			currentItemContent.HTML += c.HTML
			currentItemContent.Graph += c.Graph
		}
	}
	return
}
