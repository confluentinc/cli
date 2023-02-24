package components

import (
	"fmt"
	"log"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/confluentinc/flink-sql-client/autocomplete"
)

var LivePrefixState struct {
	LivePrefix string
	IsEnabled  bool
}

var lastStatement = ""
var allStatements = ""

func executor(in string) {
	if strings.HasSuffix(in, ";") {
		lastStatement = lastStatement + in
		LivePrefixState.IsEnabled = false
		LivePrefixState.LivePrefix = in
		allStatements = allStatements + lastStatement
		lastStatement = ""

		if isInputClosingSelect(in) {
			LivePrefixState.IsEnabled = true
			LivePrefixState.LivePrefix = ""
		}

		return
	}
	lastStatement = lastStatement + in + " "
	LivePrefixState.LivePrefix = ""
	LivePrefixState.IsEnabled = true
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnabled
}
func isInputClosingSelect(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), "SELECT") && input[len(input)-1] == ';'
}

func promptInput(value string, history []string, getSmartCompletion func() bool, toggleSmartCompletion func(), toggleOutputMode func(), exitApplication func()) (string, []string) {
	completerWithHistoryAndDocs := autocomplete.CompleterWithHistoryAndDocs(history, getSmartCompletion)

	// We need to disable the live prefix, in case we just submited a statement
	LivePrefixState.IsEnabled = false

	p := prompt.New(
		executor,
		completerWithHistoryAndDocs,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionInitialBufferText(value),
		prompt.OptionHistory(history),
		// TODO - Decide if we want to use the emacs keybind mode, or basic, or make it customizable
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
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
			Key: prompt.ControlD,
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
			Key: prompt.ControlS,
			Fn: func(b *prompt.Buffer) {
				toggleSmartCompletion()
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

func init() {
	// Print Flink's ASCII Art
	b, err := os.ReadFile("components/flink_ascii_60_with_text.txt")
	if err != nil {
		log.Printf("Couldn't read flink's ascii art. Error: %v\n", err)
	}

	// TODO - After setting up the event loop, we could maybe use tcell to get the
	// terminal's width so we disable printing the ascii art if the terminal is too small
	/* screen, _ := tcell.NewScreen()
	   screen.Init()

	   w, h := screen.Size() */
	// Right now, go-prompt get's executed first so this isn't possible
	// But we can always just delete the ascii art - it's just a gimmick

	fmt.Println(string(b))

	// Print welcome message
	fmt.Fprintf(os.Stdout, "Welcome! \033[0m%s \033[0;36m%s. \033[0m \n \n", "Flink SQL Client powered", "by Confluent")

	// Print shortcuts
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlS]", "Smart Completion ")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m \n", "[CtrlO]", "Interactive Output ON/OFF")
}

func InteractiveInput(value string, history []string, getSmartCompletion func() bool, toggleSmartCompletion func(), toggleOutputMode func(), exitApplication func()) (string, []string) {
	_, statements := promptInput(value, history, getSmartCompletion, toggleSmartCompletion, toggleOutputMode, exitApplication)

	return "", statements
}
