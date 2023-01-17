package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var input *tview.InputField

type InputController struct {
	basic func()
}

func InputControllerInit(inputRef *tview.InputField) InputController {
	input = inputRef

	basic := func() {
		table.SetBorders(false).
			SetSelectable(false, false).
			SetSeparator(' ')
	}

	input.SetDoneFunc(func(key tcell.Key) {
		selectRow()
		navigate()
	})

	return InputController{
		basic: basic,
	}
}
