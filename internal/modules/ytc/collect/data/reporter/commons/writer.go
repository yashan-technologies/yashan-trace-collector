package commons

import (
	"ytc/internal/modules/ytc/collect/resultgenner/reporter"

	"github.com/jedib0t/go-pretty/v6/table"
)

func GenStringWriter(title string, rows ...string) reporter.Writer {
	tw := ReporterWriter.NewTableWriter()
	tw.AppendHeader(table.Row{title})
	for _, row := range rows {
		tw.AppendRow(table.Row{row})
	}
	return tw
}

func GenPathWriter(path string) reporter.Writer {
	return GenStringWriter("存放路径", path)
}
