package controller

import (
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway"
	"github.com/confluentinc/flink-sql-client/components"
	"github.com/rivo/tview"
	"os"

	"github.com/gdamore/tcell/v2"
)

// Tview application.
var app = tview.NewApplication()

type ApplicationController struct {
	InitInteractiveOutputMode func()
	focus                     func(component string)
	PrintTable                func()
	fetchDataAndPrintTable    func()
	suspendOutputMode         func(f func())
	toggleSmartCompletion     func()
	toggleOutputMode          func()
	exitApplication           func()
	getView                   func() string
	getSmartCompletion        func() bool
	getOutputMode             func() string
}

func ApplicationControllerInit(store Store, tableController TableController, inputController InputController, shortcutsController ShortcutsController) ApplicationController {
	var data string
	tAppSuspended := true
	smartCompletion := true
	outputMode := "static" // interactive or static

	// Actions
	initInteractiveOutputMode := func() {
		tAppSuspended = false
	}

	suspendOutputMode := func(f func()) {
		tAppSuspended = true
		app.Suspend(f)
		tAppSuspended = false

	}

	toggleSmartCompletion := func() {
		smartCompletion = !smartCompletion
	}

	toggleOutputMode := func() {
		if outputMode == "interactive" {
			outputMode = "static"
		} else {
			outputMode = "interactive"
		}
	}

	focus := func(component string) {
		switch component {
		case "table":
			tableController.focus()
		}
	}

	printTable := func() {
		tableController.setData(data)
	}

	fetchDataAndPrintTable := func() {
		data = store.fetchData()
		printTable()
	}

	// This function should be used to proparly stop the application, cache saving, cleanup and so on
	exitApplication := func() {
		SaveHistory(inputController.getHistory())
		app.Stop()
		os.Exit(0)
	}

	// Getters
	getView := func() string {
		if tAppSuspended {
			return "inputMode"
		} else {
			return "outputMode"
		}
	}

	getOutputMode := func() string {
		return outputMode
	}

	getSmartCompletion := func() bool {
		return smartCompletion
	}

	// Function to handle shortcuts and keybindings for the whole app
	appInputCapture := func(tableController TableController) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlQ {
				exitApplication()
				return nil
			} else if event.Key() == tcell.KeyCtrlT {
				tableController.borders()
				return nil
				// TODO we have to actually go forward and backwards and not only go to the next mock
			} else if event.Key() == tcell.KeyCtrlN || event.Key() == tcell.KeyCtrlP {
				fetchDataAndPrintTable()
				return nil
			} else if event.Key() == tcell.KeyCtrlC {
				if !table.HasFocus() {
					// TODO move this to appController stop
					exitApplication()
				} else {
					tableController.onCtrlC()
				}
				return nil
			}
			return event
		}
	}

	// Set Input Capture for the whole application
	app.SetInputCapture(appInputCapture(tableController))

	return ApplicationController{initInteractiveOutputMode, focus, printTable, fetchDataAndPrintTable, suspendOutputMode, toggleSmartCompletion, toggleOutputMode, exitApplication, getView, getSmartCompletion, getOutputMode}
}

func StartApp() {
	store := NewStore()

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
	inputController.RunInteractiveInput()
	appController.InitInteractiveOutputMode()
	interactiveOutput := components.InteractiveOutput(input, table, shortcuts)
	appController.PrintTable()

	rootLayout := components.RootLayout(interactiveOutput)
	// TODO: configure and use
	gatewayClient := v2.NewAPIClient(&v2.Configuration{})
	fmt.Printf("gateway client config %T", gatewayClient.GetConfig())

	// Start the application.
	if err := app.SetRoot(rootLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
