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

type ApplicationControllerInterface interface {
	SuspendOutputMode(callback func())
	ToggleOutputMode()
	GetOutputMode() OutputMode
	ExitApplication()
	TView() *tview.Application
}

type ApplicationController struct {
	app        *tview.Application
	outputMode OutputMode
	history    *History
}

type ApplicationOptions struct {
	MOCK_STATEMENTS_OUTPUT_DEMO bool
	HTTP_CLIENT_UNSAFE_TRACE    bool
	FLINK_GATEWAY_URL           string
	DEFAULT_PROPERTIES          map[string]string
}

var once sync.Once

func (a *ApplicationController) SuspendOutputMode(cb func()) {
	a.ToggleOutputMode()
	a.app.Suspend(cb)
	// InteractiveInput has already set the data to be displayed in the table
	// Now we just need to render it
	a.app.ForceDraw()
}

func (a *ApplicationController) ToggleOutputMode() {
	if a.outputMode == TViewOutput {
		a.outputMode = GoPromptOutput
	} else {
		a.outputMode = TViewOutput
	}
}

// This function should be used to proparly stop the application, cache saving, cleanup and so on
func (a *ApplicationController) ExitApplication() {
	a.history.Save()
	a.app.Stop()
	os.Exit(0)
}

func (a *ApplicationController) GetOutputMode() OutputMode {
	return a.outputMode
}

func (a *ApplicationController) TView() *tview.Application {
	return a.app
}

func NewApplicationController(app *tview.Application, history *History) ApplicationControllerInterface {
	return &ApplicationController{
		app:        app,
		outputMode: TViewOutput,
		history:    history,
	}
}

func StartApp(envId, orgResourceId, kafkaClusterId, computePoolId, authToken string, authenticated func() error, appOptions *ApplicationOptions) {
	// Client used to communicate with the gateway
	client := NewGatewayClient(envId, orgResourceId, kafkaClusterId, computePoolId, authToken, appOptions)

	// Load history of previous commands from cache file
	history := LoadHistory()

	// Create Components
	table := components.CreateTable()
	shortcuts := components.Shortcuts()
	app := tview.NewApplication()

	// Instantiate Application Controller - this is the top level controller that will be passed down to all other controllers
	// and should be used for functions that are not specific to a component
	appController := NewApplicationController(app, history)

	// Store used to process statements and store local properties
	store := NewStore(client, appOptions, appController)

	// Instantiate Component Controllers
	tableController := NewTableController(table, store, appController)
	inputController := NewInputController(tableController, appController, store, authenticated, history, appOptions)
	shortcutsController := NewShortcutsController(shortcuts, appController, tableController)

	// Pass input controller to table controller - input and output view interact with each other and that it easier
	tableController.SetInputController(inputController)

	// Event handlers
	app.SetInputCapture(tableController.AppInputCapture)
	shortcuts.SetHighlightedFunc(shortcutsController.ShortcutHighlighted)
	interactiveOutput := components.InteractiveOutput(table, shortcuts)
	rootLayout := components.RootLayout(interactiveOutput)

	// We start tview and then suspend it immediately so we intialize all components
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if !screen.HasPendingEvent() {
			once.Do(func() {
				go appController.SuspendOutputMode(inputController.RunInteractiveInput)
			})
		}
	})

	// Start the application.
	if err := appController.TView().SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
