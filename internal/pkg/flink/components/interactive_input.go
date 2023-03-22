package components

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed flink_ascii_60_with_text.txt
var flinkAsciiArt []byte

var LivePrefixState struct {
	LivePrefix string
	IsEnabled  bool
}

var LastStatement = ""
var AllStatements = ""

func Executor(in string) {
	if strings.HasSuffix(in, ";") {
		LastStatement = LastStatement + in
		LivePrefixState.IsEnabled = false
		LivePrefixState.LivePrefix = in
		AllStatements = AllStatements + LastStatement
		LastStatement = ""

		if IsInputClosingSelect(in) {
			LivePrefixState.IsEnabled = true
			LivePrefixState.LivePrefix = ""
		}

		return
	}
	LastStatement = LastStatement + in + " "
	LivePrefixState.LivePrefix = ""
	LivePrefixState.IsEnabled = true
}

func printEmptySpaces(n int) {
	for i := 0; i < n; i++ {
		fmt.Print(" ")
	}
}

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
	printEmptySpaces(maxCol - len(prefix+stateMsg))
	fmt.Print("\n")
}
func ChangeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnabled
}
func IsInputClosingSelect(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), "SELECT") && input[len(input)-1] == ';'
}

func init() {
	// TODO - check terminal's width so we disable printing the ascii art if the terminal is too small
	// we can use tview or go-prompt for this. Either GetMaxCol from inputController or use tcell like this:
	/* screen, _ := tcell.NewScreen()
	   screen.Init()

	   w, h := screen.Size() */

	fmt.Println(string(flinkAsciiArt))

	// Print welcome message
	fmt.Fprintf(os.Stdout, "Welcome! \033[0m%s \033[0;36m%s. \033[0m \n \n", "Flink SQL Client powered", "by Confluent")

	// Print shortcuts
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlS]", "Smart Completion")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m \n", "[CtrlO]", "Interactive Output")
}
