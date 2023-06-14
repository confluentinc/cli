package controller

import (
	"os"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

// Tview application.

type TableMode string

var (
	PlaintextTable   TableMode = "plaintext"
	InteractiveTable TableMode = "interactive"
)

type ApplicationController struct {
	app              *tview.Application
	outputMode       types.OutputMode
	history          *history.History
	tableView        tview.Primitive
	cleanupFunctions []func()
}

// preFunc will, if defined, before the main function is executed. Both are executed after tview is suspended.
func (a *ApplicationController) SuspendOutputMode(callback func()) {
	a.ToggleOutputMode()
	a.app.Suspend(callback)
	// InteractiveInput has already set the data to be displayed in the table
	// Now we just need to render it
	a.app.ForceDraw()
}

func (a *ApplicationController) ToggleOutputMode() {
	if a.outputMode == types.TViewOutput {
		a.outputMode = types.GoPromptOutput
	} else {
		a.outputMode = types.TViewOutput
	}
}

// This function should be used to proparly stop the application, cache saving, cleanup and so on
func (a *ApplicationController) ExitApplication() {
	for _, cleanupFunction := range a.cleanupFunctions {
		cleanupFunction()
	}
	a.history.Save()
	a.app.Stop()
	os.Exit(0)
}

func (a *ApplicationController) GetOutputMode() types.OutputMode {
	return a.outputMode
}

func (a *ApplicationController) TView() *tview.Application {
	return a.app
}

func (a *ApplicationController) StartTView(layout tview.Primitive) error {
	a.tableView = layout
	return a.app.SetRoot(layout, true).EnableMouse(false).Run()
}

func (a *ApplicationController) ShowTableView() {
	a.app.SetRoot(a.tableView, true).EnableMouse(false)
}

func (a *ApplicationController) AddCleanupFunction(cleanupFunction func()) types.ApplicationControllerInterface {
	a.cleanupFunctions = append(a.cleanupFunctions, cleanupFunction)
	return a
}

func NewApplicationController(app *tview.Application, history *history.History) types.ApplicationControllerInterface {
	return &ApplicationController{
		app:        app,
		outputMode: types.TViewOutput,
		history:    history,
	}
}
