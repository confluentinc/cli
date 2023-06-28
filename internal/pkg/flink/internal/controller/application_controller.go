package controller

import (
	"os"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Tview application.

type TableMode string

type ApplicationController struct {
	app              *tview.Application
	history          *history.History
	tableView        tview.Primitive
	cleanupFunctions []func()
}

func (a *ApplicationController) SuspendOutputMode() {
	a.app.Stop()
}

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
		log.CliLogger.Errorf("Failed to open table tview., %v", err)
		utils.OutputErr("Error: failed to open table tview")
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
