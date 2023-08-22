package reporter

import (
	"fmt"

	"ytc/utils/stringutil"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	_CSS_CLASS_TABLE = "ytc_table"
	_CSS_CLASS_LIST  = "ytc_list"
)

const _PLACEHOLDER = "--"

// validate interface
var _ Writer = (table.Writer)(nil)
var _ Writer = (list.Writer)(nil)
var _ Writer = (*ErrorWriter)(nil)

type ErrorWriter struct {
	listWriter list.Writer
}

type Writer interface {
	Render() string
	RenderMarkdown() string
	RenderHTML() string
}

type ReportWriter struct{}

func NewReporterWriter() *ReportWriter { return &ReportWriter{} }

// NewListWriter returns a table writer with default stype: StyleRounded.
func (rw *ReportWriter) NewTableWriter(style ...table.Style) table.Writer {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	if len(style) > 0 {
		tw.SetStyle(style[0])
	}
	tw.Style().Title = table.TitleOptions{
		Align:  text.AlignCenter,
		Format: text.FormatUpper,
	}
	tw.Style().HTML = table.HTMLOptions{
		CSSClass:    _CSS_CLASS_TABLE,
		EmptyColumn: "&nbsp;",
		EscapeText:  false,
		Newline:     "<br/>",
	}
	return tw
}

// NewListWriter returns a list writer with default stype: StyleBulletCircle.
func (rw *ReportWriter) NewListWriter(style ...list.Style) list.Writer {
	lw := list.NewWriter()
	lw.SetStyle(list.StyleBulletCircle)
	if len(style) > 0 {
		lw.SetStyle(style[0])
	}
	lw.SetHTMLCSSClass(_CSS_CLASS_LIST)
	return lw
}

// NewErrorWriter returns a error writer contains error data.
func (rw *ReportWriter) NewErrorWriter(err, description string) *ErrorWriter {
	// trans empty string to placeholder
	if stringutil.IsEmpty(err) {
		err = _PLACEHOLDER
	}
	if stringutil.IsEmpty(description) {
		description = _PLACEHOLDER
	}

	// append data to listWriter
	lw := rw.NewListWriter()
	lw.AppendItem(fmt.Sprintf("报错：%s", err))
	lw.AppendItem(fmt.Sprintf("描述：%s", description))

	return &ErrorWriter{listWriter: lw}
}

// [Interface Func]
func (ew *ErrorWriter) Render() string {
	return ew.listWriter.Render()
}

// [Interface Func]
func (ew *ErrorWriter) RenderMarkdown() string {
	return ew.listWriter.RenderMarkdown()
}

// [Interface Func]
func (ew *ErrorWriter) RenderHTML() string {
	return ew.listWriter.RenderHTML()
}
