package controller

import (
	"github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway"
	"github.com/confluentinc/flink-sql-client/components"
	"github.com/rivo/tview"
	"os"

	"github.com/gdamore/tcell/v2"
)

// Tview application.

type ApplicationController struct {
	tAppSuspended       bool
	app                 *tview.Application
	smartCompletion     bool
	outputMode          string
	store               Store
	inputController     *InputController
	shortcutsController ShortcutsController
	tableController     *TableController
}

func (a *ApplicationController) getView() string {
	if a.tAppSuspended {
		return "inputMode"
	} else {
		return "outputMode"
	}
}
func (a *ApplicationController) InitInteractiveOutputMode() {
	a.tAppSuspended = false
}

func (a *ApplicationController) suspendOutputMode(cb func()) {
	a.tAppSuspended = true
	a.app.Suspend(cb)
	a.tAppSuspended = false
}

func (a *ApplicationController) toggleSmartCompletion() {
	a.smartCompletion = !a.smartCompletion
}

func (a *ApplicationController) toggleOutputMode() {
	if a.outputMode == "interactive" {
		a.outputMode = "static"
	} else {
		a.outputMode = "interactive"
	}
}

func (a *ApplicationController) focus(component string) {
	switch component {
	case "table":
		a.tableController.focus()
	}
}

func (a *ApplicationController) PrintTable(data string) {
	a.tableController.setData(data)
}

func (a *ApplicationController) fetchDataAndPrintTable() {
	data := a.store.fetchData()
	a.PrintTable(data)
}

// This function should be used to proparly stop the application, cache saving, cleanup and so on
func (a *ApplicationController) exitApplication() {
	a.inputController.History.Save()
	a.app.Stop()
	os.Exit(0)
}

func (a *ApplicationController) getOutputMode() string {
	return a.outputMode
}

func (a *ApplicationController) getSmartCompletion() bool {
	return a.smartCompletion
}

// Function to handle shortcuts and keybindings for the whole app
func (a *ApplicationController) appInputCapture(tableController *TableController) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			a.exitApplication()
			return nil
		} else if event.Key() == tcell.KeyCtrlT {
			tableController.borders()
			return nil
			// TODO we have to actually go forward and backwards and not only go to the next mock
		} else if event.Key() == tcell.KeyCtrlN || event.Key() == tcell.KeyCtrlP {
			a.fetchDataAndPrintTable()
			return nil
		} else if event.Key() == tcell.KeyCtrlC {
			if !tableController.table.HasFocus() {
				// TODO move this to appController stop
				a.exitApplication()
			} else {
				tableController.onCtrlC()
			}
			return nil
		}
		return event
	}
}

func NewApplicationController(store Store, tableController *TableController, inputController *InputController, shortcutsController ShortcutsController) ApplicationController {

	// Set Input Capture for the whole application

	controller := ApplicationController{
		store:               store,
		inputController:     inputController,
		shortcutsController: shortcutsController,
		tableController:     tableController,
		app:                 tview.NewApplication(),
		smartCompletion:     true,
		outputMode:          "static",
		tAppSuspended:       true,
	}
	controller.app.SetInputCapture(controller.appInputCapture(tableController))
	return controller
}

func StartApp() {
	store := NewStore(v2.NewAPIClient(&v2.Configuration{}))

	// Create Components
	table := components.CreateTable()
	input := components.InputField()
	shortcuts := components.Shortcuts()

	// Instantiate Component Controllers
	tableController := NewTableController(table)
	inputController := NewInputController(input)
	shortcutsController := NewShortcutsController(shortcuts)

	// Instatiate Application Controller
	appController := NewApplicationController(store, &tableController, &inputController, shortcutsController)
	tableController.appController = &appController
	inputController.appController = &appController
	tableController.InputController = &inputController
	shortcutsController.appController = &appController
	shortcutsController.tableController = &tableController

	// Event handlers
	input.SetDoneFunc(inputController.onDone)
	input.SetInputCapture(inputController.HandleKeyEvent)
	table.SetInputCapture(tableController.handleCellEvent)
	shortcuts.SetHighlightedFunc(shortcutsController.shortcutHighlighted)

	// Instantiate interactive components
	inputController.RunInteractiveInput()
	appController.InitInteractiveOutputMode()
	interactiveOutput := components.InteractiveOutput(input, table, shortcuts)

	rootLayout := components.RootLayout(interactiveOutput)
	// TODO: configure and use

	// Start the application.
	if err := appController.app.SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
