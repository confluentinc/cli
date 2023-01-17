package main

import (
	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/rivo/tview"
)

// Tview application.
var app = tview.NewApplication()

func main() {
	// Instantiate Controllers and Components
	tableController := TableControllerInit(components.CreateTable())
	ShortcutsControllerInit(components.Shortcuts(), tableController)
	InputControllerInit(components.InputField())
	ApplicationControllerInit(tableController)

	// Instantiate Interactive Components
	components.InteractiveInput()
	interactiveOutput := components.InteractiveOutput(input, table)

	// Create the main layout.
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(interactiveOutput, 0, 1, true).
		AddItem(shortcuts, 1, 1, false)

	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
