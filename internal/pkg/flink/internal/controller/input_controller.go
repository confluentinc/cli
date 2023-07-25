package controller

import (
	"errors"
	"reflect"
	"strings"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/autocomplete"
	lexer "github.com/confluentinc/cli/internal/pkg/flink/internal/highlighting"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/reverseisearch"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type InputController struct {
	History               *history.History
	InitialBuffer         string
	smartCompletion       bool
	reverseISearchEnabled bool
	prompt                prompt.IPrompt
	shouldExit            bool
	reverseISearch        reverseisearch.ReverseISearch
}

const defaultWindowSize = 100

func NewInputController(history *history.History) types.InputControllerInterface {
	inputController := &InputController{
		History:         history,
		InitialBuffer:   "",
		smartCompletion: true,
		shouldExit:      false,
		reverseISearch:  reverseisearch.NewReverseISearch(),
	}
	inputController.prompt = inputController.Prompt()
	return inputController
}

func (c *InputController) GetUserInput() string {
	// if the initial buffer is not empty, we insert the text and reset the InitialBuffer
	if c.InitialBuffer != "" {
		c.prompt.Buffer().InsertText(c.InitialBuffer, false, true)
		c.InitialBuffer = ""
	}
	return c.prompt.Input()
}

func (c *InputController) HasUserInitiatedExit(userInput string) bool {
	// the user input should actually never be an empty string. The only case in which go-prompt returns an empty string,
	// is when the user presses CtrlD. This is why we need to specifically handle this case here.
	userPressedCtrlD := userInput == ""
	if c.shouldExit || userPressedCtrlD {
		return true
	}
	return false
}

func (c *InputController) HasUserEnabledReverseSearch() bool {
	return c.reverseISearchEnabled
}

func (c *InputController) StartReverseSearch() {
	searchResult := c.reverseISearch.ReverseISearch(c.History.Data)
	c.reverseISearchEnabled = false
	c.InitialBuffer = searchResult
}

func (c *InputController) GetWindowWidth() int {
	windowSize, err := c.getMaxCol()
	if err != nil {
		return defaultWindowSize
	}
	return windowSize
}

// This function fetches the current max column width for the terminal
// In other words, the amount of characters that can be displayed in one line
func (c *InputController) getMaxCol() (int, error) {
	p := c.prompt
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Pointer {
		return -1, errors.New("could not reflect prompt")
	} else {
		v = v.Elem()
	}

	v = v.FieldByName("renderer")
	if v.Kind() != reflect.Pointer {
		return -1, errors.New("could not reflect prompt.renderer")
	} else {
		v = v.Elem()
	}

	v = v.FieldByName("col")
	if v.Kind() != reflect.Uint16 {
		return -1, errors.New("could not reflect prompt.renderer.col")
	}

	maxCol := v.Uint()

	return int(maxCol), nil
}

func (c *InputController) Prompt() prompt.IPrompt {
	completer := autocomplete.NewCompleterBuilder(c.getSmartCompletion).
		AddCompleter(autocomplete.ExamplesCompleter).
		AddCompleter(autocomplete.SetCompleter).
		AddCompleter(autocomplete.ShowCompleter).
		AddCompleter(autocomplete.GenerateHistoryCompleter(c.History.Data)).
		BuildCompleter()

	return prompt.New(
		nil,
		completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory(c.History.Data),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionSetExitCheckerOnInput(func(input string, breakline bool) bool {
			return c.reverseISearchEnabled || c.shouldExit
		}),
		prompt.OptionAddASCIICodeBind(),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlD,
			Fn: func(b *prompt.Buffer) {
				c.shouldExit = true
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn: func(b *prompt.Buffer) {
				c.shouldExit = true
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlS,
			Fn: func(b *prompt.Buffer) {
				c.toggleSmartCompletion()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlR,
			Fn: func(b *prompt.Buffer) {
				c.reverseISearchEnabled = true
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
		prompt.OptionSetLexer(lexer.Lexer),
		prompt.OptionSetStatementTerminator(func(lastKeyStroke prompt.Key, buffer *prompt.Buffer) bool {
			text := buffer.Text()
			text = strings.TrimSpace(text)
			// We add exit here because we also want to exit without the need of adding semicolon, which is the default flow for all statements
			if text == "exit" {
				return true
			}
			if len(text) == 0 || text[len(text)-1] != ';' {
				return false
			}
			return true
		}),
	)
}

func (c *InputController) getSmartCompletion() bool {
	return c.smartCompletion
}

func (c *InputController) toggleSmartCompletion() {
	c.smartCompletion = !c.smartCompletion

	maxCol, err := c.getMaxCol()
	if err != nil {
		log.CliLogger.Error(err)
		return
	}

	components.PrintSmartCompletionState(c.getSmartCompletion(), maxCol)
}
