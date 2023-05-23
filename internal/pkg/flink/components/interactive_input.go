package components

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

func PrintSmartCompletionState(smartCompletion bool, maxCol int) {
	msgPrefix := "Smart Completion is now "
	PrintOptionState(msgPrefix, smartCompletion, maxCol)
}

func PrintOutputModeState(outputMode bool, maxCol int) {
	msgPrefix := "Interactive output is now "
	PrintOptionState(msgPrefix, outputMode, maxCol)
}

func PrintOptionState(prefix string, state bool, maxCol int) {
	stateMsg := "disabled"
	if state {
		stateMsg = "enabled"
	}

	fmt.Fprintf(os.Stdout, "\n\033[0m%s\033[0;36m%s\033[0m", prefix, stateMsg)

	// This prints to the console the exact amount of empty characters to fill the line might have autocompletions before
	strings.Repeat(" ", maxCol-len(prefix+stateMsg))
	fmt.Print("\n")
}

func IsInputClosingSelect(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), "SELECT") && input[len(input)-1] == ';'
}

// This prints flinks ascii art, welcome message and shortcuts
func PrintWelcomeHeader() {
	// Print welcome message
	fmt.Fprintf(os.Stdout, "Welcome! \n")

	// Print shortcuts
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m \n", "[CtrlS]", "Smart Completion")
	// disabled for now
	//fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m \n", "[CtrlO]", "Interactive Output")
}
