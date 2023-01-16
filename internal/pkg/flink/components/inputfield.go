package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputField demonstrates the InputField.
func InputField(nextSlide func(), app *tview.Application) (title string, params ExtraSlideParams, content tview.Primitive) {
	input := tview.NewInputField().
		SetLabel("Enter a number: ").
		SetAcceptanceFunc(tview.InputFieldInteger).SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	return "Input", ExtraSlideParams{}, tview.NewFlex().
		AddItem(input, 300, 1, true)
}
