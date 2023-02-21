package main

import (
	"fmt"
	v2 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway"
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
	shortcutsController := ShortcutsControllerInit(shortcuts, tableController, &appController)

	// Instatiate Application Controller
	appController = ApplicationControllerInit(store, tableController, inputController, shortcutsController)

	// Instantiate interactive components
	inputController.runInteractiveInput()
	appController.initInteractiveOutputMode()
	interactiveOutput := components.InteractiveOutput(input, table, shortcuts)
	appController.printTable()

	rootLayout := components.RootLayout(interactiveOutput)
	// TODO: configure and use
	gatewayClient := v2.NewAPIClient(&v2.Configuration{})
	fmt.Printf("gateway client config %T", gatewayClient.GetConfig())

	// Start the application.
	if err := app.SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
