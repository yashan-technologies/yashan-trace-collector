package collect

import (
	"fmt"
	"os"
	"strings"
	constdef "ytc/defs/constants"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/stringutil"
	"ytc/utils/terminalutil"
)

const (
	SAVE = "Save"
	QUIT = "Quit"

	FORM_HEADER = "Enter Yashan Trace Collecter Data"
)

func (c *CollectCmd) openYasdbCollectForm() (*yasdb.YasdbEnv, int) {
	var opts []terminalutil.WithOption
	opts = append(opts, func(c *terminalutil.CollectFrom) {
		c.AddInput(constdef.YASDB_HOME, os.Getenv(constdef.YASDB_HOME), validatePath)
	})
	opts = append(opts, func(c *terminalutil.CollectFrom) {
		c.AddInput(constdef.YASDB_DATA, os.Getenv(constdef.YASDB_DATA), validatePath)
	})
	opts = append(opts, func(c *terminalutil.CollectFrom) { c.AddInput(constdef.YASDB_USER, "", nil) })
	opts = append(opts, func(c *terminalutil.CollectFrom) { c.AddPassword(constdef.YASDB_PASSWORD, "", nil) })
	opts = append(opts, func(c *terminalutil.CollectFrom) { c.AddButton(SAVE, saveFunc) })
	opts = append(opts, func(c *terminalutil.CollectFrom) { c.AddButton(QUIT, quitFunc) })
	form := terminalutil.NewCollectFrom(FORM_HEADER, opts...)
	form.Start()
	yasdbEnv, err := getYasdbEnvFromForm(form)
	if err != nil {
		return nil, terminalutil.FormExitNotContinue
	}
	return yasdbEnv, form.ExitCode
}

func getYasdbEnvFromForm(c *terminalutil.CollectFrom) (*yasdb.YasdbEnv, error) {
	yasdbHome, err := c.GetFormData(constdef.YASDB_HOME)
	if err != nil {
		return nil, err
	}
	yasdbData, err := c.GetFormData(constdef.YASDB_DATA)
	if err != nil {
		return nil, err
	}
	yasdbUser, err := c.GetFormData(constdef.YASDB_USER)
	if err != nil {
		return nil, err
	}
	yasdbPassword, err := c.GetFormData(constdef.YASDB_PASSWORD)
	if err != nil {
		return nil, err
	}
	return &yasdb.YasdbEnv{
		YasdbHome:     trimSpace(yasdbHome),
		YasdbData:     trimSpace(yasdbData),
		YasdbUser:     trimSpace(yasdbUser),
		YasdbPassword: trimSpace(yasdbPassword),
	}, nil
}

func validatePath(label, value string) (bool, string) {
	if stringutil.IsEmpty(value) {
		return false, fmt.Sprintf("please enter %s", label)
	}
	if _, err := os.Stat(value); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func saveFunc(c *terminalutil.CollectFrom) {
	log.Controller.Debugf("exec internal")
	if err := c.Validate(); err != nil {
		c.ShowTips(err.Error())
		return
	}
	// TODO : 对话框不能剧中显示
	// yasdbEnv, err := getYasdbEnvFromForm(c)
	// log.Controller.Debugf("get yasdb env %s ,%s", jsonutil.ToJSONString(yasdbEnv), jsonutil.ToJSONString(err))
	// if err != nil {
	// 	c.Stop(terminalutil.FormExitNotContinue)
	// }
	// log.Controller.Debugf("start validate password")
	// items := []string{
	// 	data.BASE_YASDB_PARAMTER,
	// 	data.DIAG_YASDB_ADR,
	// 	data.DIAG_YASDB_RUNLOG,
	// }
	// var buffer bytes.Buffer
	// if err := yasdbEnv.ValidYasdbUserAndPwd(); err != nil {
	// 	buffer.WriteString(err.Error() + "\n\n")
	// 	for i, item := range items {
	// 		buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, collectWaringMsg(item)))
	// 	}
	// 	buffer.WriteString("\nClick OK you will enter next start collect\nClick Cancel you can continue full table")
	// 	c.ConfrimExit(buffer.String())
	// 	return
	// }
	c.Stop(terminalutil.FormExitContinue)
}

func quitFunc(c *terminalutil.CollectFrom) {
	c.Stop(terminalutil.FormExitNotContinue)
}

// func collectWaringMsg(item string) string {
// 	return fmt.Sprintf("%s will be collect failed", item)
// }

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
