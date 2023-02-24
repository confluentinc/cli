package controller

import (
	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InputController struct {
	lastStatement string
	statements    []string
	History       History
	appController *ApplicationController
	input         *tview.InputField
}

// Actions
// This will be run after tview.app gets suspended
// Upon returning tview.app will be resumed.
func (c *InputController) RunInteractiveInput() {
	run := func() {
		// This prints again the last fetched data as a raw text table to the inputMode
		if c.lastStatement != "" && c.appController.getOutputMode() == "interactive" {
			c.appController.fetchDataAndPrintTable()
		}

		// Executed after tview.app is suspended and before go-prompt takes over
		c.lastStatement = c.input.GetText()

		// Run interactive input and take over terminal
		c.lastStatement, c.statements = components.InteractiveInput(c.lastStatement, c.History.Data, c.appController.getSmartCompletion, c.appController.toggleSmartCompletion, c.appController.toggleOutputMode, c.appController.exitApplication)

		// Executed still while tview.app is suspended and after go-prompt has finished
		c.input.SetText(c.lastStatement)
		c.History.Append(c.statements)
	}

	// Run interactive input, take over terminal and save output to lastStatement and statements
	run()

	for c.appController.getOutputMode() == "static" {
		c.appController.fetchDataAndPrintTable()
		run()
	}
}

func (c *InputController) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		func() {
			c.appController.suspendOutputMode(c.RunInteractiveInput)
			c.appController.fetchDataAndPrintTable()
		}()

		return nil
	}

	return event
}

func (c *InputController) onDone(key tcell.Key) {
	c.appController.focus("table")
}

func NewInputController(inputRef *tview.InputField) (c InputController) {
	// Variables
	c.input = inputRef
	// Initialization
	c.History = LoadHistory()
	return c
}
