package controller

import (
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway"
	"github.com/confluentinc/flink-sql-client/components"
	"github.com/rivo/tview"
	"os"
)

// Tview application.

type TableMode string

var (
	PlaintextTable   TableMode = "plaintext"
	InteractiveTable TableMode = "interactive"
)

type OutputMode string

var (
	GoPromptOutput OutputMode = "goprompt"
	TViewOutput    OutputMode = "tview"
)

type ApplicationController struct {
	tAppSuspended bool
	app           *tview.Application
	outputMode    OutputMode
	history       History
}

func (a *ApplicationController) getMode() TableMode {
	if a.tAppSuspended {
		return PlaintextTable
	} else {
		return InteractiveTable
	}
}
func (a *ApplicationController) InitInteractiveOutputMode() {
	a.tAppSuspended = false
}

func (a *ApplicationController) suspendOutputMode(cb func()) {
	a.toggleOutputMode()
	a.app.Suspend(cb)
	a.toggleOutputMode()
}

func (a *ApplicationController) toggleOutputMode() {
	if a.outputMode == TViewOutput {
		a.outputMode = GoPromptOutput
	} else {
		a.outputMode = TViewOutput
	}
}

// This function should be used to proparly stop the application, cache saving, cleanup and so on
func (a *ApplicationController) exitApplication() {
	a.history.Save()
	a.app.Stop()
	os.Exit(0)
}

func (a *ApplicationController) getOutputMode() OutputMode {
	return a.outputMode
}

func NewApplicationController(app *tview.Application, history History) *ApplicationController {
	return &ApplicationController{
		app:           app,
		outputMode:    GoPromptOutput,
		tAppSuspended: true,
		history:       history,
	}
}

func StartApp() {
	store := NewStore(v2.NewAPIClient(&v2.Configuration{}))
	history := LoadHistory()
	// Create Components
	table := components.CreateTable()
	shortcuts := components.Shortcuts()
	app := tview.NewApplication()

	// Instantiate Component Controllers
	appController := NewApplicationController(app, history)
	tableController := NewTableController(table, store, appController)

	inputController := NewInputController(history, tableController, appController, store)
	shortcutsController := NewShortcutsController(shortcuts, appController, tableController)

	// Instatiate Application Controller
	tableController.InputController = &inputController

	// Event handlers
	table.SetInputCapture(tableController.handleCellEvent)
	app.SetInputCapture(tableController.appInputCapture)
	shortcuts.SetHighlightedFunc(shortcutsController.shortcutHighlighted)

	// Instantiate interactive components
	inputController.RunInteractiveInput()
	appController.InitInteractiveOutputMode()
	interactiveOutput := components.InteractiveOutput(table, shortcuts)

	rootLayout := components.RootLayout(interactiveOutput)
	// TODO: configure and use

	// Start the application.
	if err := appController.app.SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
