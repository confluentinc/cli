package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
)

type ApplicationController struct {
	focus             func(component string)
	printTable        func()
	fetchData         func()
	suspendOutputMode func(f func())
	getView           func() string
}

var quit = func() {
	app.Stop()
	os.Exit(0)
}

func ApplicationControllerInit(store Store, tableController TableController, inputController InputController, shortcutsController ShortcutsController) ApplicationController {
	var data string
	tAppSuspended := false

	// Actions
	suspendOutputMode := func(f func()) {
		tAppSuspended = true
		app.Suspend(f)
		tAppSuspended = false
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

	getView := func() string {
		if tAppSuspended {
			return "inputMode"
		} else {
			return "outputMode"
		}
	}

	// Function to handle shortcuts and keybindings for the whole app
	appInputCapture := func(tableController TableController) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlQ {
				quit()
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
					quit()
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

	return ApplicationController{focus, printTable, fetchDataAndPrintTable, suspendOutputMode, getView}
}
