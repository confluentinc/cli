package components

import (
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/confluentinc/flink-sql-client/autocomplete"
)

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

var lastStatement = ""
var allStatements = ""

func executor(in string) {
	if strings.HasSuffix(in, ";") {
		lastStatement = lastStatement + in
		LivePrefixState.IsEnable = false
		LivePrefixState.LivePrefix = in
		allStatements = allStatements + lastStatement
		lastStatement = ""
		return
	}
	lastStatement = lastStatement + in + " "
	LivePrefixState.LivePrefix = ""
	LivePrefixState.IsEnable = true
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}
func isInputClosingSelect(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), "SELECT") && input[len(input)-1] == ';'
}

func promptInput(value string, history []string, toggleOutputMode func(), exitApplication func()) (string, []string) {
	completerWithHistory := autocomplete.CompleterWithHistory(history)

	p := prompt.New(
		executor,
		completerWithHistory,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionInitialBufferText(value),
		prompt.OptionHistory(history),
		// TODO - Decide if we want to use the emacs keybind mode, or basic, or make it customizable
		prompt.SwitchKeyBindMode(prompt.CommonKeyBind),
		prompt.OptionSetExitCheckerOnInput(func(input string, breakline bool) bool {
			if input == "" {
				return false
			} else if isInputClosingSelect(input) && breakline {
				return true
			} else {
				return false
			}
		}),
		prompt.OptionAddASCIICodeBind(),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(b *prompt.Buffer) {
				exitApplication()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn: func(b *prompt.Buffer) {
				exitApplication()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlO,
			Fn: func(b *prompt.Buffer) {
				toggleOutputMode()
			},
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x62},
			Fn:        prompt.GoLeftWord,
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x66},
			Fn:        prompt.GoRightWord,
		}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionSetLexer(lexer),
	)

	p.Run()

	// We need to remove the trailing empty string from the split
	var statements = strings.Split(allStatements, ";")
	statements = statements[:len(statements)-1]
	lastStatement = statements[len(statements)-1]
	return lastStatement, statements
}

func printPrefix() {
	fmt.Print("Flink SQL Client \n")
	// The escape sequences below are used to color the text.
	// Not all colors are compatible with all terminals
	// Here is a stackoverflow with more details https://stackoverflow.com/questions/4842424/list-of-ansi-color-escape-sequences
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlS]", "Smart Completion ")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m \n \n", "[CtrlO]", "Interactive Output ON/OFF")
}

func InteractiveInput(value string, history []string, toggleOutputMode func(), exitApplication func()) (string, []string) {
	printPrefix()
	fmt.Print("flinkSQL")
	var lastStatement, statements = promptInput(value, history, toggleOutputMode, exitApplication)

	return lastStatement, statements
}
