package components

import (
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

var lastStatement = ""
var allStatements = ""

func completer(in prompt.Document) []prompt.Suggest {

	s := []prompt.Suggest{
		{Text: "SELECT", Description: "Select data from a database"},
		{Text: "INSERT", Description: "Add rows to a table"},
		{Text: "DESCRIBE", Description: "Describe the schema of a table or a view"},
		{Text: "SET", Description: "Set current database or catalog"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

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

func promptInput(value string) (string, []string) {
	prompt.NewStdoutWriter().WriteRawStr("completer")

	p := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionInitialBufferText(value),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.SwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionSetExitCheckerOnInput(func(input string, breakline bool) bool {
			if input == "" {
				return false
			} else if isInputClosingSelect(input) && breakline {
				return true
			} else {
				return false
			}
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x62},
			Fn:        prompt.GoLeftWord,
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x66},
			Fn:        prompt.GoRightWord,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn: func(b *prompt.Buffer) {
				os.Exit(0)
			},
		}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionLivePrefix(changeLivePrefix),
	)

	p.Run()

	// We need to remove the trailing empty string from the split
	var statements = strings.Split(allStatements, ";")
	statements = statements[:len(statements)-1]
	finalStatement := statements[len(statements)-1]
	return finalStatement, statements
}

func printPrefix() {
	fmt.Print("Flink SQL Client \n")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;93m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;93m%s \033[0m", "[CtrlS]", "Smart Completion ")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;93m%s \033[0m \n \n", "[CtrlM]", "Interactive Output ON/OFF")
}

func InteractiveInput(value string) string {
	printPrefix()
	fmt.Print("flinkSQL")
	var in, _ = promptInput(value)

	return in
}
