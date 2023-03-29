package controller

import (
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/confluentinc/flink-sql-client/components"
	"github.com/rivo/tview"
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
	app        *tview.Application
	outputMode OutputMode
	history    *History
}

type ApplicationOptions struct {
	MOCK_STATEMENTS_OUTPUT_DEMO bool
}

var once sync.Once

func (a *ApplicationController) suspendOutputMode(cb func()) {
	a.toggleOutputMode()
	a.app.Suspend(cb)
	// InteractiveInput has already set the data to be displayed in the table
	// Now we just need to render it
	a.app.ForceDraw()
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

func NewApplicationController(app *tview.Application, history *History) *ApplicationController {
	return &ApplicationController{
		app:        app,
		outputMode: TViewOutput,
		history:    history,
	}
}

func StartApp(envId, computePoolId, authToken string, appOptions *ApplicationOptions) {
	client := NewGatewayClient(envId, computePoolId, authToken)
	store := NewStore(client, appOptions)
	history := LoadHistory()
	// Create Components
	table := components.CreateTable()
	shortcuts := components.Shortcuts()
	app := tview.NewApplication()

	// Instantiate Component Controllers
	appController := NewApplicationController(app, history)
	tableController := NewTableController(table, store, appController)

	inputController := NewInputController(tableController, appController, store, history)
	shortcutsController := NewShortcutsController(shortcuts, appController, tableController)

	// Instatiate Application Controller
	tableController.InputController = &inputController

	// Event handlers
	table.SetInputCapture(tableController.handleCellEvent)
	app.SetInputCapture(tableController.appInputCapture)
	shortcuts.SetHighlightedFunc(shortcutsController.shortcutHighlighted)
	interactiveOutput := components.InteractiveOutput(table, shortcuts)
	rootLayout := components.RootLayout(interactiveOutput)

	// We start tview and then suspend it immediately so we intialize all components
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if !screen.HasPendingEvent() {
			once.Do(func() {
				go appController.suspendOutputMode(inputController.RunInteractiveInput)
			})
		}
	})

	// Start the application.
	if err := appController.app.SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
