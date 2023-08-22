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
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaserr"
	"github.com/jedib0t/go-pretty/v6/table"
)

// validate interface
var _ commons.Reporter = (*HostCPUUsageReporter)(nil)

type HostCPUUsageReporter struct{}

type sarCPUUsage struct {
	timestamp int64
	sar.CpuUsage
}

type gopsutilCPUUsage struct {
	timestamp int64
	gopsutil.CpuUsage
}

func NewHostCPUUsageReporter() HostCPUUsageReporter {
	return HostCPUUsageReporter{}
}

// [Interface Func]
func (r HostCPUUsageReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2
	txt := reporter.GenTxtTitle(title)
	markdown := reporter.GenMarkdownTitle(title, fontSize)
	html := reporter.GenHTMLTitle(title, fontSize)

	historyItem, currentItem, err := validateWorkLoadItem(item)
	if err != nil {
		err = yaserr.Wrapf(err, "validate cpu usage item")
		return
	}

	historyItemContent, err := r.genHistoryContent(historyItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate cpu usage history content")
		return
	}

	currentItemContent, err := r.genCurrentContent(currentItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate cpu usage current content")
		return
	}

	content.Txt = strings.Join([]string{txt, historyItemContent.Txt, currentItemContent.Txt}, stringutil.STR_NEWLINE)
	content.Markdown = strings.Join([]string{markdown, historyItemContent.Markdown, currentItemContent.Markdown}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{html, historyItemContent.HTML, currentItemContent.HTML}, stringutil.STR_NEWLINE)
	return
}

func (r HostCPUUsageReporter) parseSarItem(item datadef.YTCItem) (output map[int64]map[string]sar.CpuUsage, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal sar cpu usage")
		return
	}
	output = make(map[int64]map[string]sar.CpuUsage)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal sar cpu usage")
		return
	}
	return
}

func (r HostCPUUsageReporter) parseSarHistoryItem(historyItem datadef.YTCItem) (output map[int64]map[string]sar.CpuUsage, err error) {
	output, err = r.parseSarItem(historyItem)
	if err != nil {
		err = yaserr.Wrapf(err, "history sar item")
	}
	return
}

func (r HostCPUUsageReporter) parseSarCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]sar.CpuUsage, err error) {
	output, err = r.parseSarItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current sar item")
	}
	return
}

func (r HostCPUUsageReporter) parseGopsutilItem(item datadef.YTCItem) (output map[int64]map[string]gopsutil.CpuUsage, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal gopsutil cpu usage")
		return
	}
	output = make(map[int64]map[string]gopsutil.CpuUsage)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal gopsutil cpu usage")
		return
	}
	return
}

func (r HostCPUUsageReporter) parseGopsutilCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]gopsutil.CpuUsage, err error) {
	output, err = r.parseGopsutilItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current gopsutil item")
	}
	return
}

func (r HostCPUUsageReporter) genSarReportContent(sarData map[int64]map[string]sar.CpuUsage) (content reporter.ReportContent) {
	tmp := make(map[string][]sarCPUUsage)
	for time, val := range sarData {
		for k, v := range val {
			cpuUsage := sarCPUUsage{
				timestamp: time,
				CpuUsage:  v,
			}
			tmp[k] = append(tmp[k], cpuUsage)
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
		"用户空间进程所花费的CPU时间(user)",
		"以较低优先级运行的用户空间进程所花费的CPU时间(nice)",
		"内核空间进程所花费的CPU时间(system)",
		"CPU等待磁盘 I/O 完成的时间(iowait)",
		"发生CPU窃取的时间(steal)",
		"CPU处于空闲状态的时间(idle)",
	})
	for _, key := range keys {
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				fmt.Sprintf("%.2f%%", p.User),
				fmt.Sprintf("%.2f%%", p.Nice),
				fmt.Sprintf("%.2f%%", p.System),
				fmt.Sprintf("%.2f%%", p.IOWait),
				fmt.Sprintf("%.2f%%", p.Steal),
				fmt.Sprintf("%.2f%%", p.Idle),
			})
		}
		c := reporter.GenReportContentByWriter(tw)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		tw.ResetRows()
	}
	return
}

func (r HostCPUUsageReporter) genGopsutilReportContent(gopsutilData map[int64]map[string]gopsutil.CpuUsage) (content reporter.ReportContent) {
	tmp := make(map[string][]gopsutilCPUUsage)
	for time, val := range gopsutilData {
		for k, v := range val {
			cpuUsage := gopsutilCPUUsage{
				timestamp: time,
				CpuUsage:  v,
			}
			tmp[k] = append(tmp[k], cpuUsage)
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
		"用户空间进程所花费的CPU时间(user)",
		"以较低优先级运行的用户空间进程所花费的CPU时间(nice)",
		"内核空间进程所花费的CPU时间(system)",
		"CPU等待磁盘 I/O 完成的时间(iowait)",
		"发生CPU窃取的时间(steal)",
		"CPU处于空闲状态的时间(idle)",
		"处理硬件中断的时间",
		"处理软件中断的时间",
		"运行虚拟机所花费的CPU时间",
	})
	for _, key := range keys {
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				fmt.Sprintf("%.2f%%", p.User),
				fmt.Sprintf("%.2f%%", p.Nice),
				fmt.Sprintf("%.2f%%", p.System),
				fmt.Sprintf("%.2f%%", p.Iowait),
				fmt.Sprintf("%.2f%%", p.Steal),
				fmt.Sprintf("%.2f%%", p.Idle),
				fmt.Sprintf("%.2f%%", p.Irq),
				fmt.Sprintf("%.2f%%", p.Softirq),
				fmt.Sprintf("%.2f%%", p.Guest),
			})
		}
		c := reporter.GenReportContentByWriter(tw)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		tw.ResetRows()
	}
	return
}

func (r HostCPUUsageReporter) genHistoryContent(historyItem datadef.YTCItem, titlePrefix string) (historyItemContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.1 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_HISTORY])
	fontSize := reporter.FONT_SIZE_H3
	if len(historyItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(historyItem.Error, historyItem.Description)
		historyItemContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		historyItemContent = reporter.GenReportContentByTitle(title, fontSize)
		history, e := r.parseSarHistoryItem(historyItem)
		if e != nil {
			err = yaserr.Wrapf(e, "parse history cpu usage")
			return
		}
		c := r.genSarReportContent(history)
		historyItemContent.Txt += c.Txt
		historyItemContent.Markdown += c.Markdown
		historyItemContent.HTML += c.HTML
	}
	return
}

func (r HostCPUUsageReporter) genCurrentContent(currentItem datadef.YTCItem, titlePrefix string) (currentItemContent reporter.ReportContent, err error) {
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
				err = yaserr.Wrapf(e, "parse sar current cpu usage")
				return
			}
			c := r.genSarReportContent(current)
			currentItemContent.Txt += c.Txt
			currentItemContent.Markdown += c.Markdown
			currentItemContent.HTML += c.HTML
		} else {
			current, e := r.parseGopsutilCurrentItem(currentItem)
			if e != nil {
				err = yaserr.Wrapf(e, "parse gopsutil current cpu usage")
				return
			}
			c := r.genGopsutilReportContent(current)
			currentItemContent.Txt += c.Txt
			currentItemContent.Markdown += c.Markdown
			currentItemContent.HTML += c.HTML
		}
	}
	return
}
