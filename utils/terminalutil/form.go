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
	ok := tview.NewButton("OK").SetSelectedFunc(func() {
		f.Stop(FormExitContinue)
	})
	cancel := tview.NewButton("Cancel").SetSelectedFunc(func() {
		f.app.SetRoot(f.form, false)
	})
	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(ok, 0, 1, false).
		AddItem(cancel, 0, 1, false)
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText(errMsg).SetTextAlign(tview.AlignLeft), 0, 1, false).
		AddItem(buttonFlex, 0, 1, false)
	box := tview.NewBox()
	box.SetRect(0, 0, 50, 10)
	box.SetBorder(true).SetTitle("Warning")
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		flex.Draw(screen)
		return x, y, width, height
	})
	f.app.SetRoot(flex, true)
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

type LeftAlignModal struct {
	*tview.Modal
	textView *tview.TextView
}

func NewLeftAlignModal() *LeftAlignModal {
	modal := &LeftAlignModal{
		Modal:    tview.NewModal(),
		textView: tview.NewTextView().SetTextAlign(tview.AlignLeft),
	}
	modal.SetBackgroundColor(tcell.ColorRed)
	return modal
}

func (m *LeftAlignModal) Draw(screen tcell.Screen) {
	m.Modal.Draw(screen)

	x, y, width, height := m.GetRect()
	m.textView.SetRect(x+1, y+1, width-2, height-4)
	m.textView.Draw(screen)
}
