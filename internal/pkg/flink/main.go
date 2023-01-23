package main

import (
	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/rivo/tview"
)

// Tview application.
var app = tview.NewApplication()

func main() {
	// Create temp store
	store := store()

	// Create Components
	table := components.CreateTable()
	input := components.InputField()
	shortcuts := components.Shortcuts()

	// Instantiate Component Controllers
	var appController ApplicationController
	tableController := TableControllerInit(table, &appController)
	inputController := InputControllerInit(input, &appController)
	shortcutsController := ShortcutsControllerInit(shortcuts, tableController)

	// Instatiate Application Controller
	appController = ApplicationControllerInit(store, tableController, inputController, shortcutsController)

	// Instantiate interactive components
	inputController.runInteractiveInput()
	interactiveOutput := components.InteractiveOutput(input, table, shortcuts)

	rootLayout := components.RootLayout(interactiveOutput)

	// Start the application.
	if err := app.SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
