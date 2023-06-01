package app

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func StartApp(client ccloudv2.GatewayClientInterface, authenticated func() error, appOptions types.ApplicationOptions) {
	var once sync.Once

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
	store := store.NewStore(client, appController.ExitApplication, &appOptions)

	// Instantiate Component Controllers
	tableController := controller.NewTableController(table, store, appController)
	inputController := controller.NewInputController(tableController, appController, store, authenticated, history, &appOptions)

	// Pass RunInteractiveInputFunc to table controller so the user can come back from the output view
	tableController.SetRunInteractiveInputCallback(inputController.RunInteractiveInput)

	// Event handlers
	app.SetInputCapture(tableController.AppInputCapture)

	interactiveOutput := components.InteractiveOutput(table, shortcuts)
	rootLayout := components.RootLayout(interactiveOutput)

	// We start tview and then suspend it immediately so we initialize all components
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if !screen.HasPendingEvent() {
			once.Do(func() {
				go appController.SuspendOutputMode(inputController.RunInteractiveInput)
			})
		}
	})

	// Start the application.
	if err := appController.StartTView(rootLayout); err != nil {
		panic(err)
	}
}
