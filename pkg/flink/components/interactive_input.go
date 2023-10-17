package components

import (
	"strings"

	fColor "github.com/fatih/color"

	"github.com/confluentinc/cli/v3/pkg/color"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func PrintSmartCompletionState(smartCompletion bool, maxCol int) {
	msgPrefix := "Smart Completion is now "
	PrintOptionState(msgPrefix, smartCompletion, maxCol)
}

func PrintOptionState(prefix string, isEnabled bool, maxCol int) {
	stateMsg := "disabled"
	if isEnabled {
		stateMsg = "enabled"
	}

	output.Print(false, "\n"+prefix+fColor.CyanString(stateMsg))

	// This prints to the console the exact amount of empty characters to fill the line might have autocompletions before
	output.Println(false, strings.Repeat(" ", maxCol-len(prefix+stateMsg)))
}

func PrintWelcomeHeader() {
	// Print welcome message
	output.Print(false, "Welcome! \n")
	output.Print(false, "To exit, press Ctrl-Q or type \"exit\". \n\n")

	// Print shortcuts
	c := fColor.New(color.AccentColor)
	output.Printf(false, "[Ctrl-Q] %s [Ctrl-S] %s \n", c.Sprint("Quit"), c.Sprint("Toggle Smart Completion"))
}
