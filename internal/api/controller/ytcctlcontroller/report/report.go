package report

type ReportCmd struct {
	Input  string `name:"input"  short:"i" help:"The collection result input."`
	Type   string `name:"type"   short:"t" help:"Type of report generated, choose one from (txt)."`
	Output string `name:"output" short:"o" help:"The output dir of the report."`
}

// [Interface Func]
func (c ReportCmd) Run() error {
	return nil
}
