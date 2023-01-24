package main

import (
	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var input *tview.InputField

type InputController struct {
	runInteractiveInput             func()
	supendOutputModeAndRunInputMode func()
}

func InputControllerInit(inputRef *tview.InputField, appController *ApplicationController) InputController {
	// Variables
	input = inputRef
	var value = ""

	// Actions
	// This will be run after tview.App gets suspended
	// Upon returning tview.App will be resumed.
	runInteractiveInput := func() {
		r := func() {
			// This prints again the last fetched data as a raw text table to the inputMode
			if value != "" {
				appController.printTable()
			}

			// Executed after tview.App is suspended and before go-prompt takes over
			value = input.GetText()

			// Run interactive input and take over terminal
			value = components.InteractiveInput(value, appController.toggleOutputMode)

			// Executed still while tview.App is suspended and after go-prompt has finished
			input.SetText(value)
		}

		r()

		if appController.getOutputMode() == "static" {
			appController.fetchDataAndPrintTable()
			r()
		}
	}

	supendOutputModeAndRunInputMode := func() {
		appController.suspendOutputMode(runInteractiveInput)
		appController.fetchDataAndPrintTable()
	}

	// Event handlers
	input.SetDoneFunc(func(key tcell.Key) {
		appController.focus("table")
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			supendOutputModeAndRunInputMode()

			return nil
		}

		return event
	})

	return InputController{
		runInteractiveInput:             runInteractiveInput,
		supendOutputModeAndRunInputMode: supendOutputModeAndRunInputMode,
	}
}
