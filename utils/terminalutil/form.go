package terminalutil

import (
	"errors"

	"ytc/defs/errdef"
	"ytc/log"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	FORM_EXIT_CONTINUE     = 0
	FORM_EXIT_NOT_CONTINUE = 1
	FORM_EXIT_CTRLC        = 2

	CONTINUE = "Continue"
	BACK     = "Back"
)

type WithOption func(c *CollectForm)

type CollectForm struct {
	app            *tview.Application
	form           *tview.Form
	inputFields    []*formInput
	inputPasswords []*formPwssword
	buttons        []*formButton
	header         string
	ExitCode       int
}

type formInput struct {
	label        string                                   // input名称
	defaultValue string                                   // 获取改label默认值
	validFunc    func(label, value string) (bool, string) // 校验是否通过，不通过的提示信息
}

type formPwssword struct {
	label        string                                   // input名称
	defaultValue string                                   // 获取改label默认值
	validFunc    func(label, value string) (bool, string) // 校验是否通过，不通过的提示信息
}

type formButton struct {
	label string
	click func(c *CollectForm)
}

func NewCollectFrom(header string, opts ...WithOption) *CollectForm {
	form := &CollectForm{
		app:            tview.NewApplication(),
		form:           tview.NewForm(),
		inputFields:    make([]*formInput, 0),
		inputPasswords: make([]*formPwssword, 0),
		header:         header,
	}
	for _, opt := range opts {
		opt(form)
	}
	return form
}

func (f *CollectForm) AddInput(label string, defaultValue string, validateFunc func(string, string) (bool, string)) {
	f.inputFields = append(f.inputFields, &formInput{
		label:        label,
		defaultValue: defaultValue,
		validFunc:    validateFunc,
	})
}

func (f *CollectForm) AddPassword(label string, defaultValue string, validateFunc func(string, string) (bool, string)) {
	f.inputPasswords = append(f.inputPasswords, &formPwssword{
		label:        label,
		defaultValue: defaultValue,
		validFunc:    validateFunc,
	})
}

func (f *CollectForm) AddButton(buttonName string, click func(c *CollectForm)) {
	f.form.AddButton(buttonName, func() {
		click(f)
	})
}

func (f *CollectForm) GetFormData(label string) (string, error) {
	keyFormItem := f.form.GetFormItemByLabel(label)
	if keyFormItem == nil {
		log.Controller.Errorf("get data by %s err :%s", label)
		return "", errdef.NewFormItemUnfound(label)
	}
	return keyFormItem.(*tview.InputField).GetText(), nil
}

func (f *CollectForm) GetFormDataByLabels(labels ...string) (res map[string]string, err error) {
	res = make(map[string]string)
	for _, label := range labels {
		value, valueErr := f.GetFormData(label)
		if valueErr != nil {
			err = valueErr
			return
		}
		res[label] = value
	}
	return
}

func (f *CollectForm) ShowTips(desc string) {
	modal := tview.NewModal().
		SetBackgroundColor(tcell.ColorRed).
		SetText(desc).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Ok" {
				f.app.SetRoot(f.form, false)
			}
		})
	f.app.SetRoot(modal, true)
}

func (f *CollectForm) Validate() error {
	for _, input := range f.inputFields {
		value, err := f.GetFormData(input.label)
		if err != nil {
			log.Controller.Errorf("get fields err :%s", err.Error())
			continue
		}
		if input.validFunc == nil {
			continue
		}
		log.Controller.Debugf("startting validate")
		is, desc := input.validFunc(input.label, value)
		if !is {
			return errors.New(desc)
		}
	}
	return nil
}

func (f *CollectForm) ConfrimExit(errMsg string) {
	modal := tview.NewModal().
		SetBackgroundColor(tcell.ColorRed).
		SetText(errMsg).
		AddButtons([]string{CONTINUE, BACK}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == CONTINUE {
				f.Stop(FORM_EXIT_CONTINUE)
				return
			}
			if buttonLabel == BACK {
				f.app.SetRoot(f.form, false)
				return
			}
		})
	f.app.SetRoot(modal, true)
}

func (f *CollectForm) Stop(code int) {
	f.ExitCode = code
	f.app.Stop()
}

func (f *CollectForm) Start() {
	for _, field := range f.inputFields {
		f.form.AddInputField(field.label, field.defaultValue, 100, nil, nil)
	}
	for _, field := range f.inputPasswords {
		f.form.AddPasswordField(field.label, field.defaultValue, 100, '*', nil)
	}
	for _, b := range f.buttons {
		f.form.AddButton(b.label, func() {
			b.click(f)
		})
	}
	f.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			f.ExitCode = FORM_EXIT_CTRLC
		}
		return event
	})
	f.form.SetBorder(true).SetTitle(f.header).SetTitleAlign(tview.AlignLeft)
	if err := f.app.SetRoot(f.form, true).EnableMouse(true).Run(); err != nil {
		log.Controller.Errorf("start yasdb collect form err :%s", err.Error())
		panic(err)
	}
}
