package app

import (
	"sync"

	"github.com/confluentinc/flink-sql-client/components"
	"github.com/confluentinc/flink-sql-client/internal/controller"
	"github.com/confluentinc/flink-sql-client/internal/history"
	"github.com/confluentinc/flink-sql-client/internal/store"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var once sync.Once

func StartApp(envId, orgResourceId, kafkaClusterId, computePoolId, authToken string, authenticated func() error, appOptions *types.ApplicationOptions) {
	// Client used to communicate with the gateway
	client := store.NewGatewayClient(envId, orgResourceId, kafkaClusterId, computePoolId, authToken, appOptions)

	// Load history of previous commands from cache file
	history := history.LoadHistory()

	// Create Components
	table := components.CreateTable()
	shortcuts := components.Shortcuts()
	app := tview.NewApplication()

	// Instantiate Application Controller - this is the top level controller that will be passed down to all other controllers
	// and should be used for functions that are not specific to a component
	appController := controller.NewApplicationController(app, history)

	// Store used to process statements and store local properties
	store := store.NewStore(client, appController.ExitApplication, appOptions)

	// Instantiate Component Controllers
	tableController := controller.NewTableController(table, store, appController)
	inputController := controller.NewInputController(tableController, appController, store, authenticated, history, appOptions)
	shortcutsController := controller.NewShortcutsController(shortcuts, appController, tableController)

	// Pass RunInteractiveInputFunc to table controller so the user can come back from the output view
	tableController.SetRunInteractiveInputCallback(inputController.RunInteractiveInput)

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
