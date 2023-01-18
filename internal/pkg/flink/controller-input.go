package main

import (
	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var input *tview.InputField

type InputController struct {
	runInteractiveInput func()
}

func InputControllerInit(inputRef *tview.InputField, appController *ApplicationController) InputController {
	// Variables
	input = inputRef
	var value = ""

	// Actions
	runInteractiveInput := func() {
		value = input.GetText()
		value = components.InteractiveInput(value)
		input.SetText(value)
	}

	supendAndRunInteractiveInput := func() {
		app.Suspend(runInteractiveInput)
	}

	// Event handlers
	input.SetDoneFunc(func(key tcell.Key) {
		appController.focus("table")
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			supendAndRunInteractiveInput()
			return nil
		}

		return event
	})

	return InputController{
		runInteractiveInput: runInteractiveInput,
	}
}
