package controller

import (
	"os"

	"github.com/confluentinc/flink-sql-client/internal/history"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/rivo/tview"
)

// Tview application.

type TableMode string

var (
	PlaintextTable   TableMode = "plaintext"
	InteractiveTable TableMode = "interactive"
)

type ApplicationControllerInterface interface {
	SuspendOutputMode(callback func())
	ToggleOutputMode()
	GetOutputMode() types.OutputMode
	ExitApplication()
	TView() *tview.Application
	ShowTableView()
	StartTView(layout tview.Primitive) error
}

type ApplicationController struct {
	app        *tview.Application
	outputMode types.OutputMode
	history    *history.History
	tableView  tview.Primitive
}

func (a *ApplicationController) SuspendOutputMode(cb func()) {
	a.ToggleOutputMode()
	a.app.Suspend(cb)
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

func NewApplicationController(app *tview.Application, history *history.History) ApplicationControllerInterface {
	return &ApplicationController{
		app:        app,
		outputMode: types.TViewOutput,
		history:    history,
	}
}
