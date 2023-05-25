package components

import (
	"strings"

	"github.com/confluentinc/cli/internal/pkg/output"
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

	output.Printf("\n\033[0m%s\033[0;36m%s\033[0m", prefix, stateMsg)

	// This prints to the console the exact amount of empty characters to fill the line might have autocompletions before
	output.Println(strings.Repeat(" ", maxCol-len(prefix+stateMsg)))
}

func PrintWelcomeHeader() {
	// Print welcome message
	output.Printf("Welcome! \n")
	output.Printf("You can press CtrlQ or type exit; to quit the client. \n\n")

	// Print shortcuts
	output.Printf("\033[0m%s \033[0;36m%s \033[0m", "[CtrlQ]", "Quit")
	output.Printf("\033[0m%s \033[0;36m%s \033[0m \n", "[CtrlS]", "Smart Completion")
}
