package controller

import (
	"errors"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/confluentinc/flink-sql-client/autocomplete"
	components "github.com/confluentinc/flink-sql-client/components"
	"github.com/olekukonko/tablewriter"
)

type InputController struct {
	statements      []string
	History         History
	appController   *ApplicationController
	smartCompletion bool
	table           *TableController
	p               *prompt.Prompt
}

// Actions
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
	c.appController.toggleOutputMode()

	maxCol, err := c.GetMaxCol()
	if err != nil {
		log.Println(err)
		return
	}

	components.PrintOutputModeState(c.appController.getOutputMode() == TViewOutput, maxCol)
}

func (c *InputController) RunInteractiveInput() {

	for c.appController.getOutputMode() == GoPromptOutput {
		// Run interactive input and take over terminal
		input := c.p.Input()

		c.History.Append([]string{input})
		if c.appController.getOutputMode() == GoPromptOutput {
			data := c.table.store.FetchData(input)
			// Print raw text table
			printResultToSTDOUT(data)
		}
	}

	// If output mode is TViewOutput we display the interactive table
	if c.appController.outputMode == TViewOutput && c.appController.tAppSuspended {
		c.table.fetchDataAndPrintTable()
	}
}

func printResultToSTDOUT(data string) {
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetHeader([]string{"OrderDate", "Region", "Rep", "Item", "Units", "UnitCost", "Total"})

	for _, tableRow := range strings.Split(data, "\n") {
		row := strings.Split(tableRow, "|")
		rawTable.Append(row)
	}

	rawTable.Render() // Send output
}

func (c *InputController) Prompt() *prompt.Prompt {
	completerWithDocsExamples := autocomplete.CompleterWithDocsExamples(c.getSmartCompletion)

	// We need to disable the live prefix, in case we just submited a statement
	components.LivePrefixState.IsEnabled = false

	return prompt.New(
		components.Executor,
		completerWithDocsExamples,
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
				c.appController.exitApplication()
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn: func(b *prompt.Buffer) {
				c.appController.exitApplication()
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
			if len(text) > 0 && text[len(text)-1] != ';' {
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

func NewInputController(history History, t *TableController, a *ApplicationController) (c InputController) {
	// Initialization
	c.History = history
	c.smartCompletion = true
	c.table = t
	c.appController = a
	c.p = c.Prompt()
	return c
}
