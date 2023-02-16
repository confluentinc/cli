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
	getHistory                      func() []string
}

func InputControllerInit(inputRef *tview.InputField, appController *ApplicationController) InputController {
	// Variables
	input = inputRef
	var lastStatement, statements = "", []string{}

	// Initialization
	history := loadHistory()

	// Actions
	// This will be run after tview.App gets suspended
	// Upon returning tview.App will be resumed.
	runInteractiveInput := func() {
		run := func() {
			// This prints again the last fetched data as a raw text table to the inputMode
			if lastStatement != "" && appController.getOutputMode() == "interactive" {
				appController.printTable()
			}

			// Executed after tview.App is suspended and before go-prompt takes over
			lastStatement = input.GetText()

			// Run interactive input and take over terminal
			lastStatement, statements = components.InteractiveInput(lastStatement, history, appController.getSmartCompletion, appController.toggleSmartCompletion, appController.toggleOutputMode, appController.exitApplication)

			// Executed still while tview.App is suspended and after go-prompt has finished
			input.SetText(lastStatement)
			history = appendToHistory(history, statements)
		}

		// Run interactive input, take over terminal and save output to lastStatement and statements
		run()

		for appController.getOutputMode() == "static" {
			appController.fetchDataAndPrintTable()
			run()
		}
	}

	supendOutputModeAndRunInputMode := func() {
		appController.suspendOutputMode(runInteractiveInput)
		appController.fetchDataAndPrintTable()
	}

	// Getters
	getHistory := func() []string {
		return history
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
		getHistory:                      getHistory,
	}
}
