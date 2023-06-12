package components

import (
	"strings"

	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/fatih/color"
)

func PrintSmartCompletionState(smartCompletion bool, maxCol int) {
	msgPrefix := "Smart Completion is now "
	PrintOptionState(msgPrefix, smartCompletion, maxCol)
}

func PrintOutputModeState(outputMode bool, maxCol int) {
	msgPrefix := "Interactive output is now "
	PrintOptionState(msgPrefix, outputMode, maxCol)
}

func PrintOptionState(prefix string, isEnabled bool, maxCol int) {
	stateMsg := "disabled"
	if isEnabled {
		stateMsg = "enabled"
	}

	output.Printf("\n" + prefix + color.CyanString(stateMsg))

	// This prints to the console the exact amount of empty characters to fill the line might have autocompletions before
	output.Println(strings.Repeat(" ", maxCol-len(prefix+stateMsg)))
}

func PrintWelcomeHeader() {
	// Print welcome message
	output.Printf("Welcome! \n")
	output.Printf("To exit, press Ctrl-Q or type \"exit\". \n\n")

	// Print shortcuts
	output.Printf("[Ctrl-Q] %s [Ctrl-S] %s \n", color.CyanString("Quit"), color.CyanString("Toggle Smart Completion"))
}
