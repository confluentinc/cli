package controller

import (
	"os"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

// Tview application.

type TableMode string

type ApplicationController struct {
	app              *tview.Application
	history          *history.History
	tableView        tview.Primitive
	cleanupFunctions []func()
}

// preFunc will, if defined, before the main function is executed. Both are executed after tview is suspended.
func (a *ApplicationController) SuspendOutputMode() {
	a.app.Stop()
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

func (a *ApplicationController) TView() *tview.Application {
	return a.app
}

func (a *ApplicationController) SetLayout(layout tview.Primitive) {
	a.tableView = layout
	a.app.SetRoot(a.tableView, true).EnableMouse(false)
}

func (a *ApplicationController) StartTView() {
	err := a.app.Run()
	if err != nil {
		panic(err)
	}
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
		app:     app,
		history: history,
	}
}
