package main

import (
	"github.com/gdamore/tcell/v2"
)

type ApplicationController struct {
}

var appInputCapture = func(tableController TableController) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlT {
			tableController.borders()
			return nil
		}
		return event
	}
}

func ApplicationControllerInit(tableController TableController, inputController InputController, shortcutsController ShortcutsController) ApplicationController {

	// Set Input Capture for the whole application
	app.SetInputCapture(appInputCapture(tableController))

	return ApplicationController{}
}
