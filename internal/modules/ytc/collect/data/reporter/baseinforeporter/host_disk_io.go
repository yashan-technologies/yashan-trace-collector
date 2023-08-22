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
	"git.yasdb.com/go/yasutil/size"
	"github.com/jedib0t/go-pretty/v6/table"
)

// validate interface
var _ commons.Reporter = (*HostDiskIOReporter)(nil)

type HostDiskIOReporter struct{}

type sarDiskO struct {
	timestamp int64
	sar.DiskIO
}

type gopsutilDiskIO struct {
	timestamp int64
	gopsutil.DiskIO
}

func NewHostDiskIOReporter() HostDiskIOReporter {
	return HostDiskIOReporter{}
}

// [Interface Func]
func (r HostDiskIOReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2
	txt := reporter.GenTxtTitle(title)
	markdown := reporter.GenMarkdownTitle(title, fontSize)
	html := reporter.GenHTMLTitle(title, fontSize)

	historyItem, currentItem, err := validateWorkLoadItem(item)
	if err != nil {
		err = yaserr.Wrapf(err, "validate disk io item")
		return
	}

	historyItemContent, err := r.genHistoryContent(historyItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate disk io history content")
		return
	}

	currentItemContent, err := r.genCurrentContent(currentItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate disk io current content")
		return
	}

	content.Txt = strings.Join([]string{txt, historyItemContent.Txt, currentItemContent.Txt}, stringutil.STR_NEWLINE)
	content.Markdown = strings.Join([]string{markdown, historyItemContent.Markdown, currentItemContent.Markdown}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{html, historyItemContent.HTML, currentItemContent.HTML}, stringutil.STR_NEWLINE)
	return
}

func (r HostDiskIOReporter) parseSarItem(item datadef.YTCItem) (output map[int64]map[string]sar.DiskIO, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal sar disk io")
		return
	}
	output = make(map[int64]map[string]sar.DiskIO)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal sar disk io")
		return
	}
	return
}

func (r HostDiskIOReporter) parseSarHistoryItem(historyItem datadef.YTCItem) (output map[int64]map[string]sar.DiskIO, err error) {
	output, err = r.parseSarItem(historyItem)
	if err != nil {
		err = yaserr.Wrapf(err, "history sar item")
	}
	return
}

func (r HostDiskIOReporter) parseSarCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]sar.DiskIO, err error) {
	output, err = r.parseSarItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current sar item")
	}
	return
}

func (r HostDiskIOReporter) parseGopsutilItem(item datadef.YTCItem) (output map[int64]map[string]gopsutil.DiskIO, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal gopsutil disk io")
		return
	}
	output = make(map[int64]map[string]gopsutil.DiskIO)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal gopsutil disk io")
		return
	}
	return
}

func (r HostDiskIOReporter) parseGopsutilCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]gopsutil.DiskIO, err error) {
	output, err = r.parseGopsutilItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current gopsutil item")
	}
	return
}

func (r HostDiskIOReporter) genSarReportContent(sarData map[int64]map[string]sar.DiskIO) (content reporter.ReportContent) {
	tmp := make(map[string][]sarDiskO)
	for time, val := range sarData {
		for k, v := range val {
			diskIO := sarDiskO{
				timestamp: time,
				DiskIO:    v,
			}
			tmp[k] = append(tmp[k], diskIO)
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
		"IOPS",
		"每秒钟读取",
		"每秒钟写入",
		"平均每个请求的扇区数",
		"平均队列长度",
		"平均 I/O 请求的等待时间",
		"平均 I/O 请求的服务时间",
		"设备的利用率",
	})
	for _, key := range keys {
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				p.Tps,
				size.GenHumanReadableSize(p.RdSec*1024, 2),
				size.GenHumanReadableSize(p.WrSec*1024, 2),
				p.AvgrqSz,
				p.AvgquSz,
				fmt.Sprintf("%.2fms", p.Await),
				fmt.Sprintf("%.2fms", p.Svctm),
				fmt.Sprintf("%.2f", p.Util),
			})
		}
		c := reporter.GenReportContentByWriterAndTitle(tw, fmt.Sprintf("磁盘设备：%s", key), reporter.FONT_SIZE_H4)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		tw.ResetRows()
	}
	return
}

func (r HostDiskIOReporter) genGopsutilReportContent(gopsutilData map[int64]map[string]gopsutil.DiskIO) (content reporter.ReportContent) {
	tmp := make(map[string][]gopsutilDiskIO)
	for time, val := range gopsutilData {
		for k, v := range val {
			diskIO := gopsutilDiskIO{
				timestamp: time,
				DiskIO:    v,
			}
			tmp[k] = append(tmp[k], diskIO)
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
		"IOPS",
		"每秒钟读取",
		"每秒钟写入",
		"每秒读次数",
		"每秒写次数",
		"设备标签",
	})
	for _, key := range keys {
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				p.Iops,
				size.GenHumanReadableSize(p.KBReadSec*1024, 2),
				size.GenHumanReadableSize(p.KBWriteSec*1024, 2),
				p.ReadCountSec,
				p.WriteCountSec,
				p.Label,
			})
		}
		c := reporter.GenReportContentByWriterAndTitle(tw, fmt.Sprintf("磁盘设备：%s", key), reporter.FONT_SIZE_H4)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		tw.ResetRows()
	}
	return
}

func (r HostDiskIOReporter) genHistoryContent(historyItem datadef.YTCItem, titlePrefix string) (historyItemContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.1 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_HISTORY])
	fontSize := reporter.FONT_SIZE_H3
	if len(historyItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(historyItem.Error, historyItem.Description)
		historyItemContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		historyItemContent = reporter.GenReportContentByTitle(title, fontSize)
		history, e := r.parseSarHistoryItem(historyItem)
		if e != nil {
			err = yaserr.Wrapf(e, "parse history disk io")
			return
		}
		c := r.genSarReportContent(history)
		historyItemContent.Txt += c.Txt
		historyItemContent.Markdown += c.Markdown
		historyItemContent.HTML += c.HTML
	}
	return
}

func (r HostDiskIOReporter) genCurrentContent(currentItem datadef.YTCItem, titlePrefix string) (currentItemContent reporter.ReportContent, err error) {
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
				err = yaserr.Wrapf(e, "parse sar current disk io")
				return
			}
			c := r.genSarReportContent(current)
			currentItemContent.Txt += c.Txt
			currentItemContent.Markdown += c.Markdown
			currentItemContent.HTML += c.HTML
		} else {
			current, e := r.parseGopsutilCurrentItem(currentItem)
			if e != nil {
				err = yaserr.Wrapf(e, "parse gopsutil current disk io")
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
