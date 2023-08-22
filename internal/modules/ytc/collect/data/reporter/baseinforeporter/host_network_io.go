package baseinforeporter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ytc/defs/confdef"
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
var _ commons.Reporter = (*HostNetworkIOReporter)(nil)

type HostNetworkIOReporter struct{}

type sarNetworkIO struct {
	timestamp int64
	sar.NetworkIO
}

type gopsutilNetworkIO struct {
	timestamp int64
	gopsutil.NetworkIO
}

func NewHostNetworkIOReporter() HostNetworkIOReporter {
	return HostNetworkIOReporter{}
}

// [Interface Func]
func (r HostNetworkIOReporter) Report(item datadef.YTCItem, titlePrefix string) (content reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s %s", titlePrefix, baseinfo.BaseInfoChineseName[item.Name])
	fontSize := reporter.FONT_SIZE_H2
	txt := reporter.GenTxtTitle(title)
	markdown := reporter.GenMarkdownTitle(title, fontSize)
	html := reporter.GenHTMLTitle(title, fontSize)

	historyItem, currentItem, err := validateWorkLoadItem(item)
	if err != nil {
		err = yaserr.Wrapf(err, "validate network io content")
		return
	}

	historyItemContent, err := r.genHistoryContent(historyItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate network io history content")
		return
	}

	currentItemContent, err := r.genCurrentContent(currentItem, titlePrefix)
	if err != nil {
		err = yaserr.Wrapf(err, "generate network io current item")
		return
	}

	content.Txt = strings.Join([]string{txt, historyItemContent.Txt, currentItemContent.Txt}, stringutil.STR_NEWLINE)
	content.Markdown = strings.Join([]string{markdown, historyItemContent.Markdown, currentItemContent.Markdown}, stringutil.STR_NEWLINE)
	content.HTML = strings.Join([]string{html, historyItemContent.HTML, currentItemContent.HTML}, stringutil.STR_NEWLINE)
	return
}

func (r HostNetworkIOReporter) parseSarItem(item datadef.YTCItem) (output map[int64]map[string]sar.NetworkIO, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal sar network io")
		return
	}
	output = make(map[int64]map[string]sar.NetworkIO)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal sar network io")
		return
	}
	return
}

func (r HostNetworkIOReporter) parseSarHistoryItem(historyItem datadef.YTCItem) (output map[int64]map[string]sar.NetworkIO, err error) {
	output, err = r.parseSarItem(historyItem)
	if err != nil {
		err = yaserr.Wrapf(err, "history sar item")
	}
	return
}

func (r HostNetworkIOReporter) parseSarCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]sar.NetworkIO, err error) {
	output, err = r.parseSarItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current sar item")
	}
	return
}

func (r HostNetworkIOReporter) parseGopsutilItem(item datadef.YTCItem) (output map[int64]map[string]gopsutil.NetworkIO, err error) {
	data, err := json.Marshal(item.Details)
	if err != nil {
		err = yaserr.Wrapf(err, "marshal gopsutil network io")
		return
	}
	output = make(map[int64]map[string]gopsutil.NetworkIO)
	if err = json.Unmarshal(data, &output); err != nil {
		err = yaserr.Wrapf(err, "unmarshal gopsutil network io")
		return
	}
	return
}

func (r HostNetworkIOReporter) parseGopsutilCurrentItem(currentItem datadef.YTCItem) (output map[int64]map[string]gopsutil.NetworkIO, err error) {
	output, err = r.parseGopsutilItem(currentItem)
	if err != nil {
		err = yaserr.Wrapf(err, "current gopsutil item")
	}
	return
}

func (r HostNetworkIOReporter) genSarReportContent(sarData map[int64]map[string]sar.NetworkIO) (content reporter.ReportContent) {
	tmp := make(map[string][]sarNetworkIO)
	for time, val := range sarData {
		for k, v := range val {
			networkIO := sarNetworkIO{
				timestamp: time,
				NetworkIO: v,
			}
			tmp[k] = append(tmp[k], networkIO)
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
		"每秒钟接收的数据包数量",
		"每秒钟发送的数据包数量",
		"每秒钟接收的数据量",
		"每秒钟发送的数据量",
		"每秒钟接收的压缩数据包数量",
		"每秒钟发送的压缩数据包数量",
		"每秒钟接收的多播数据包数量",
	})
	for _, key := range keys {
		if confdef.IsDiscardNetwork(key) {
			continue
		}
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				p.Rxpck,
				p.Txpck,
				size.GenHumanReadableSize(p.RxkB*1024, 2),
				size.GenHumanReadableSize(p.TxkB*1024, 2),
				p.Rxcmp,
				p.Txcmp,
				p.Rxmcst,
			})
		}
		c := reporter.GenReportContentByWriterAndTitle(tw, fmt.Sprintf("网络接口：%s", key), reporter.FONT_SIZE_H4)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		tw.ResetRows()
	}
	return
}

func (r HostNetworkIOReporter) genGopsutilReportContent(gopsutilData map[int64]map[string]gopsutil.NetworkIO) (content reporter.ReportContent) {
	tmp := make(map[string][]gopsutilNetworkIO)
	for time, val := range gopsutilData {
		for k, v := range val {
			networkIO := gopsutilNetworkIO{
				timestamp: time,
				NetworkIO: v,
			}
			tmp[k] = append(tmp[k], networkIO)
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
		"每秒钟接收的数据包数量",
		"每秒钟发送的数据包数量",
		"每秒钟接收的数据量",
		"每秒钟发送的数据量",
		"接收过程中的错误数",
		"发送过程中的错误数",
		"被丢弃的传入数据包数",
		"被丢弃的传出数据包数",
		"接收过程中的FIFO缓冲区错误数",
		"发送过程中的FIFO缓冲区错误数",
	})
	for _, key := range keys {
		if confdef.IsDiscardNetwork(key) {
			continue
		}
		pointers := tmp[key]
		sort.Slice(pointers, func(i, j int) bool {
			return pointers[i].timestamp < pointers[j].timestamp
		})
		for _, p := range pointers {
			tw.AppendRow(table.Row{
				time.Unix(p.timestamp, 0).Format(timedef.TIME_FORMAT),
				p.Rxpck,
				p.Txpck,
				size.GenHumanReadableSize(p.RxkB*1024, 2),
				size.GenHumanReadableSize(p.TxkB*1024, 2),
				p.Errin,
				p.Errout,
				p.Dropin,
				p.Dropout,
				p.Fifoin,
				p.Fifoout,
			})
		}
		c := reporter.GenReportContentByWriterAndTitle(tw, fmt.Sprintf("网络接口：%s", key), reporter.FONT_SIZE_H4)
		content.Txt += c.Txt + stringutil.STR_NEWLINE
		content.Markdown += c.Markdown + stringutil.STR_NEWLINE
		content.HTML += c.HTML + stringutil.STR_NEWLINE
		tw.ResetRows()
	}
	return
}

func (r HostNetworkIOReporter) genHistoryContent(historyItem datadef.YTCItem, titlePrefix string) (historyItemContent reporter.ReportContent, err error) {
	title := fmt.Sprintf("%s.1 %s", titlePrefix, baseinfo.BaseInfoChildChineseName[baseinfo.KEY_HISTORY])
	fontSize := reporter.FONT_SIZE_H3
	if len(historyItem.Error) != 0 {
		ew := commons.ReporterWriter.NewErrorWriter(historyItem.Error, historyItem.Description)
		historyItemContent = reporter.GenReportContentByWriterAndTitle(ew, title, fontSize)
	} else {
		historyItemContent = reporter.GenReportContentByTitle(title, fontSize)
		history, e := r.parseSarHistoryItem(historyItem)
		if e != nil {
			err = yaserr.Wrapf(e, "parse history network io")
			return
		}
		c := r.genSarReportContent(history)
		historyItemContent.Txt += c.Txt
		historyItemContent.Markdown += c.Markdown
		historyItemContent.HTML += c.HTML
	}
	return
}

func (r HostNetworkIOReporter) genCurrentContent(currentItem datadef.YTCItem, titlePrefix string) (currentItemContent reporter.ReportContent, err error) {
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
				err = yaserr.Wrapf(e, "parse sar current network io")
				return
			}
			c := r.genSarReportContent(current)
			currentItemContent.Txt += c.Txt
			currentItemContent.Markdown += c.Markdown
			currentItemContent.HTML += c.HTML
		} else {
			current, e := r.parseGopsutilCurrentItem(currentItem)
			if e != nil {
				err = yaserr.Wrapf(e, "parse gopsutil current network io")
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
