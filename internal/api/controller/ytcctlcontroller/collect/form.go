package collect

import (
	"fmt"
	"os"
	"path"
	"strings"

	constdef "ytc/defs/constants"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/processutil"
	"ytc/utils/stringutil"
	"ytc/utils/terminalutil"
)

const (
	SAVE = "Save"
	QUIT = "Quit"

	FORM_HEADER = "Enter Yashan Trace Collector Data"

	base_yasdb_process_format       = `.*yasdb (?i:(nomount|mount|open))`
	yasdb_internal_data_not_collect = "yasdb internal data will not be collected, are you sure to continue?"
)

var (
	YasdbValidate error
)

func (c *CollectCmd) openYasdbCollectForm() (*yasdb.YasdbEnv, int) {
	yasdbHome, yasdbData := yasdbPath()
	var opts []terminalutil.WithOption
	opts = append(opts, func(c *terminalutil.CollectForm) { c.AddInput(constdef.YASDB_HOME, yasdbHome, validatePath) })
	opts = append(opts, func(c *terminalutil.CollectForm) { c.AddInput(constdef.YASDB_DATA, yasdbData, validatePath) })
	opts = append(opts, func(c *terminalutil.CollectForm) { c.AddInput(constdef.YASDB_USER, "", nil) })
	opts = append(opts, func(c *terminalutil.CollectForm) { c.AddPassword(constdef.YASDB_PASSWORD, "", nil) })
	opts = append(opts, func(c *terminalutil.CollectForm) { c.AddButton(SAVE, saveFunc) })
	opts = append(opts, func(c *terminalutil.CollectForm) { c.AddButton(QUIT, quitFunc) })
	form := terminalutil.NewCollectFrom(FORM_HEADER, opts...)
	form.Start()
	yasdbEnv, err := getYasdbEnvFromForm(form)
	if err != nil {
		return nil, terminalutil.FORM_EXIT_NOT_CONTINUE
	}
	return yasdbEnv, form.ExitCode
}

func getYasdbEnvFromForm(c *terminalutil.CollectForm) (*yasdb.YasdbEnv, error) {
	labelMap, err := c.GetFormDataByLabels(constdef.YASDB_HOME, constdef.YASDB_DATA, constdef.YASDB_USER, constdef.YASDB_PASSWORD)
	if err != nil {
		return nil, err
	}
	return &yasdb.YasdbEnv{
		YasdbHome:     trimSpace(labelMap[constdef.YASDB_HOME]),
		YasdbData:     trimSpace(labelMap[constdef.YASDB_DATA]),
		YasdbUser:     trimSpace(labelMap[constdef.YASDB_USER]),
		YasdbPassword: trimSpace(labelMap[constdef.YASDB_PASSWORD]),
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

func saveFunc(c *terminalutil.CollectForm) {
	log.Controller.Debugf("exec internal")
	if err := c.Validate(); err != nil {
		c.ShowTips(err.Error())
		return
	}
	yasdbEnv, err := getYasdbEnvFromForm(c)
	if err != nil {
		log.Controller.Errorf("get yasdb env err: %s", err.Error())
		return
	}
	if err := yasdbEnv.ValidYasdbUserAndPwd(); err != nil {
		log.Controller.Errorf("validate yasdb err: %s", err.Error())
		YasdbValidate = err
		desc, _ := ytccollectcommons.YasErrDescAndTips(err)
		desc = strings.Join([]string{desc, yasdb_internal_data_not_collect}, ", ")
		c.ConfrimExit(desc)
		return
	}
	c.Stop(terminalutil.FORM_EXIT_CONTINUE)
}

func quitFunc(c *terminalutil.CollectForm) {
	c.Stop(terminalutil.FORM_EXIT_NOT_CONTINUE)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func yasdbPath() (yasdbHome, yasdbData string) {
	yasdbData = os.Getenv(constdef.YASDB_DATA)
	yasdbHome = os.Getenv(constdef.YASDB_HOME)
	processYasdbHome, processYasdbData := yasdbPathFromProcess()
	if stringutil.IsEmpty(yasdbHome) {
		yasdbHome = processYasdbHome
	}
	if stringutil.IsEmpty(yasdbData) {
		yasdbData = processYasdbData
	}
	return
}

func yasdbPathFromProcess() (yasdbHome string, yasdbData string) {
	processes, err := processutil.ListAnyUserProcessByCmdline(base_yasdb_process_format, true)
	if err != nil {
		return
	}
	if len(processes) == 0 {
		return
	}
	for _, p := range processes {
		fields := strings.Split(p.ReadableCmdline, "-D")
		if len(fields) < 2 {
			continue
		}
		yasdbData = trimSpace(fields[1])
		full := trimSpace(p.FullCommand)
		if !path.IsAbs(full) {
			return
		}
		yasdbHome = path.Dir(path.Dir(full))
		return
	}
	return
}
