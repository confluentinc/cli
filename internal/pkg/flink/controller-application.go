package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
)

type ApplicationController struct {
	initInteractiveOutputMode func()
	focus                     func(component string)
	printTable                func()
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
	var smartCompletion = true
	var outputMode = "static" // interactive or static

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
		saveHistory(inputController.getHistory())
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
