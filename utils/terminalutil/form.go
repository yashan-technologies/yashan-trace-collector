package terminalutil

import (
	"errors"
	"ytc/defs/errdef"
	"ytc/log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	FormExitNotContinue = 1
	FormExitContinue    = 2

	Ok     = "Ok"
	Cancel = "Cancel"
)

type WithOption func(c *CollectFrom)

type CollectFrom struct {
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
	click func(c *CollectFrom)
}

func NewCollectFrom(header string, opts ...WithOption) *CollectFrom {
	form := &CollectFrom{
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

func (f *CollectFrom) AddInput(label string, defaultValue string, validateFunc func(string, string) (bool, string)) {
	f.inputFields = append(f.inputFields, &formInput{
		label:        label,
		defaultValue: defaultValue,
		validFunc:    validateFunc,
	})
}

func (f *CollectFrom) AddPassword(label string, defaultValue string, validateFunc func(string, string) (bool, string)) {
	f.inputPasswords = append(f.inputPasswords, &formPwssword{
		label:        label,
		defaultValue: defaultValue,
		validFunc:    validateFunc,
	})
}

func (f *CollectFrom) AddButton(buttonName string, click func(c *CollectFrom)) {
	f.form.AddButton(buttonName, func() {
		click(f)
	})
}

func (f *CollectFrom) GetFormData(label string) (string, error) {
	keyFormItem := f.form.GetFormItemByLabel(label)
	if keyFormItem == nil {
		log.Controller.Errorf("get data by %s err :%s", label)
		return "", errdef.NewFormItemUnfound(label)
	}
	return keyFormItem.(*tview.InputField).GetText(), nil
}

func (f *CollectFrom) ShowTips(desc string) {
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

func (f *CollectFrom) Validate() error {
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

func (f *CollectFrom) ConfrimExit(errMsg string) {
	modal := tview.NewModal().
		SetBackgroundColor(tcell.ColorRed).
		SetText(errMsg).
		AddButtons([]string{Ok, Cancel}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == Ok {
				f.Stop(FormExitContinue)
				return
			}
			if buttonLabel == Cancel {
				f.app.SetRoot(f.form, false)
				return
			}
		})
	f.app.SetRoot(modal, true)
}

func (f *CollectFrom) Stop(code int) {
	f.ExitCode = code
	f.app.Stop()
}

func (f *CollectFrom) Start() {
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
	f.form.SetBorder(true).SetTitle(f.header).SetTitleAlign(tview.AlignLeft)
	if err := f.app.SetRoot(f.form, true).EnableMouse(true).Run(); err != nil {
		log.Controller.Errorf("start yasdb collect form err :%s", err.Error())
		panic(err)
	}
}
