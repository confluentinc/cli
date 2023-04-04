package controller

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/confluentinc/flink-sql-client/autocomplete"
	components "github.com/confluentinc/flink-sql-client/components"
	"github.com/confluentinc/go-prompt"
	"github.com/olekukonko/tablewriter"
)

type InputController struct {
	History         *History
	appController   ApplicationControllerInterface
	smartCompletion bool
	table           *TableController
	p               *prompt.Prompt
	store           StoreInterface
}

// Actions
//  This is the main function/loop for the app
func (c *InputController) RunInteractiveInput() {

	var statementResult *StatementResult
	var err *StatementError
	// We check for statement result and rows so we don't leave GoPrompt in case of errors
	for c.appController.GetOutputMode() == GoPromptOutput || statementResult == nil || len(statementResult.Rows) == 0 {
		// Run interactive input and take over terminal
		input := c.p.Input()
		c.History.Append([]string{input})
		statementResult, err = c.store.ProcessStatement(input)
		renderMsgAndStatus(statementResult)
		if err != nil {
			fmt.Println(err.Error())
			c.isSessionValid(err)
			continue
		}

		if c.appController.GetOutputMode() == GoPromptOutput {
			// Print raw text table
			printResultToSTDOUT(statementResult)
		}
	}

	// If output mode is TViewOutput we set the data to be displayed in the interactive table
	if c.appController.GetOutputMode() == TViewOutput {
		c.table.setDataAndFocus(statementResult)
	}
}

func (c *InputController) isSessionValid(err *StatementError) bool {
	// exit application if user needs to authenticate again
	if err != nil && err.HttpResponseCode == http.StatusUnauthorized {
		c.appController.ExitApplication()
		return false
	}
	return true
}

func renderMsgAndStatus(statementResult *StatementResult) {
	if statementResult == nil {
		return
	}
	if statementResult.StatusDetail != "" {
		fmt.Println(statementResult.StatusDetail)
	} else {
		fmt.Println("Statement successfully submited. No details returned from server.")
	}
	if statementResult.StatementName != "" {
		fmt.Println("Statement ID: " + statementResult.StatementName)
	}
	if statementResult.Status != "" {
		fmt.Println("Current status: " + statementResult.Status + ".")
	}
}

func (c *InputController) toggleSmartCompletion() {
	c.smartCompletion = !c.smartCompletion

	maxCol, err := c.GetMaxCol()
	if err != nil {
		log.Println(err)
		return
	}

	components.PrintSmartCompletionState(c.getSmartCompletion(), maxCol)
}

func (c *InputController) toggleOutputMode() {
	c.appController.ToggleOutputMode()

	maxCol, err := c.GetMaxCol()
	if err != nil {
		log.Println(err)
		return
	}

	components.PrintOutputModeState(c.appController.GetOutputMode() == TViewOutput, maxCol)
}

func printResultToSTDOUT(data *StatementResult) {
	if len(data.Rows) == 0 {
		return
	}

	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetHeader(data.Columns)
	rawTable.AppendBulk(data.Rows)
	rawTable.Render() // Send output
}

func (c *InputController) Prompt() *prompt.Prompt {
	completer := autocomplete.NewCompleterBuilder(c.getSmartCompletion).
		AddCompleter(autocomplete.ExamplesCompleter).
		AddCompleter(autocomplete.SetCompleter).
		AddCompleter(autocomplete.ShowCompleter).
		AddCompleter(autocomplete.GenerateHistoryCompleter(&c.History.Data)).
		AddCompleter(autocomplete.GenerateDocsCompleter()).
		BuildCompleter()

	// We need to disable the live prefix, in case we just submited a statement
	components.LivePrefixState.IsEnabled = false

	return prompt.New(
		components.Executor,
		completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory(c.History.Data),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionSetExitCheckerOnInput(func(input string, breakline bool) bool {
			if input == "" {
				return false
			} else if components.IsInputClosingSelect(input) && breakline {
				return true
			} else {
				return false
			}
		}),
		prompt.OptionAddASCIICodeBind(),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlD,
			Fn: func(b *prompt.Buffer) {
				c.appController.ExitApplication()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn: func(b *prompt.Buffer) {
				c.appController.ExitApplication()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlS,
			Fn: func(b *prompt.Buffer) {
				c.toggleSmartCompletion()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlO,
			Fn: func(b *prompt.Buffer) {
				c.toggleOutputMode()
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
		prompt.OptionLivePrefix(components.ChangeLivePrefix),
		prompt.OptionSetLexer(components.Lexer),
		prompt.OptionSetStatementTerminator(func(lastKeyStroke prompt.Key, buffer *prompt.Buffer) bool {
			text := buffer.Text()
			text = strings.TrimSpace(text)
			if len(text) == 0 || text[len(text)-1] != ';' {
				return false
			}
			return true
		}),
	)
}

// Getters
func (c *InputController) getSmartCompletion() bool {
	return c.smartCompletion
}

// This function fetches the current max column width for the terminal
// In other words, the amount of characters that can be displayed in one line
func (c *InputController) GetMaxCol() (int, error) {
	p := c.p
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

func NewInputController(t *TableController, a ApplicationControllerInterface, store StoreInterface, history *History) (c InputController) {
	// Initialization
	c.History = history
	c.smartCompletion = true
	c.table = t
	c.appController = a
	c.store = store
	c.p = c.Prompt()
	components.PrintWelcomeHeader()

	return c
}
