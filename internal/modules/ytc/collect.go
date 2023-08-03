package ytc

import (
	"bytes"
	"errors"
	"os/exec"
)

func Demo(outputDir, reportType string, types map[string]struct{}) (fname string, err error) {
	cmd := exec.Command("uname", "-a")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return
	}
	if cmd.ProcessState.ExitCode() != 0 {
		err = errors.New(stderr.String())
		return
	}
	nodeDemoResult := &NodeDemoResult{
		Base: DemoBaseResult{
			Uname: stdout.String(),
		},
	}
	results := DemoResults{
		Results: map[string]*NodeDemoResult{
			"db-1-1": nodeDemoResult,
		},
	}
	return results.GenResult(outputDir, reportType, types)
}
