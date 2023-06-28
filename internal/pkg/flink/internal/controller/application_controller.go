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
	history          *history.History
	cleanupFunctions []func()
}

func (a *ApplicationController) ExitApplication() {
	for _, cleanupFunction := range a.cleanupFunctions {
		cleanupFunction()
	}
	a.history.Save()
	os.Exit(0)
}

func (a *ApplicationController) AddCleanupFunction(cleanupFunction func()) types.ApplicationControllerInterface {
	a.cleanupFunctions = append(a.cleanupFunctions, cleanupFunction)
	return a
}

func NewApplicationController(app *tview.Application, history *history.History) types.ApplicationControllerInterface {
	return &ApplicationController{
		history: history,
	}
}
