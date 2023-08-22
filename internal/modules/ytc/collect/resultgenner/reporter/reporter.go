package reporter

import (
	"fmt"
	"ytc/utils/stringutil"
)

const (
	REPORT_TYPE_TXT  ReportType = "txt"
	REPORT_TYPE_MD   ReportType = "md"
	REPORT_TYPE_HTML ReportType = "html"
)

const (
	FONT_SIZE_H1 FontSize = iota + 1
	FONT_SIZE_H2
	FONT_SIZE_H3
	FONT_SIZE_H4
)

type FontSize uint8

type ReportType string

type ReportContent struct {
	Txt      string
	Markdown string
	HTML     string
}

func GenTxtTitle(title string) string {
	return fmt.Sprintf("%s\n", title)
}

func GenMarkdownTitle(title string, fontSize ...FontSize) string {
	defaultTitle := fmt.Sprintf("## %s\n\n", title)
	if len(fontSize) > 0 {
		switch fontSize[0] {
		case FONT_SIZE_H1:
			return fmt.Sprintf("# %s\n\n", title)
		case FONT_SIZE_H2:
			return defaultTitle
		case FONT_SIZE_H3:
			return fmt.Sprintf("### %s\n\n", title)
		case FONT_SIZE_H4:
			return fmt.Sprintf("#### %s\n\n", title)
		default:
			return defaultTitle
		}
	}
	return defaultTitle
}

func GenHTMLTitle(title string, fontSize ...FontSize) string {
	defaultTitle := fmt.Sprintf("<h2>%s</h2>\n", title)
	if len(fontSize) > 0 {
		switch fontSize[0] {
		case FONT_SIZE_H1:
			return fmt.Sprintf("<h1>%s</h1>\n", title)
		case FONT_SIZE_H2:
			return defaultTitle
		case FONT_SIZE_H3:
			return fmt.Sprintf("<h3>%s</h3>\n", title)
		case FONT_SIZE_H4:
			return fmt.Sprintf("<h4>%s</h4>\n", title)
		default:
			return defaultTitle
		}
	}
	return defaultTitle
}

func GenReportContentByWriter(w Writer) ReportContent {
	return ReportContent{
		Txt:      w.Render(),
		Markdown: w.RenderMarkdown(),
		HTML:     w.RenderHTML(),
	}
}

func GenReportContentByTitle(title string, fontSize ...FontSize) ReportContent {
	return ReportContent{
		Txt:      GenTxtTitle(title),
		Markdown: GenMarkdownTitle(title, fontSize...),
		HTML:     GenHTMLTitle(title, fontSize...),
	}
}

func GenReportContentByWriterAndTitle(w Writer, title string, fontSize FontSize) ReportContent {
	writerContent := GenReportContentByWriter(w)
	titleContent := GenReportContentByTitle(title, fontSize)

	// generate txt content
	var txt string
	txt += titleContent.Txt
	txt += writerContent.Txt + stringutil.STR_NEWLINE

	// generate markdown content
	var markdown string
	markdown += titleContent.Markdown
	markdown += writerContent.Markdown + stringutil.STR_NEWLINE

	// generate html content
	var html string
	html += titleContent.HTML
	html += writerContent.HTML + stringutil.STR_NEWLINE

	return ReportContent{
		Txt:      txt,
		Markdown: markdown,
		HTML:     html,
	}
}
