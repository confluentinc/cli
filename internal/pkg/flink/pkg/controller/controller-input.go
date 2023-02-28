package controller

import (
	"github.com/c-bata/go-prompt"
	"github.com/confluentinc/flink-sql-client/autocomplete"
	components "github.com/confluentinc/flink-sql-client/components"
	"github.com/olekukonko/tablewriter"
	"os"
	"strings"
)

type InputController struct {
	statements      []string
	History         History
	appController   *ApplicationController
	smartCompletion bool
	table           *TableController
}

func (c *InputController) getSmartCompletion() bool {
	return c.smartCompletion
}

func (c *InputController) toggleSmartCompletion() {
	c.smartCompletion = !c.smartCompletion
}

// Actions
// This will be run after tview.app gets suspended
// Upon returning tview.app will be resumed.
func (c *InputController) RunInteractiveInput() {

	for c.appController.getOutputMode() == GoPromptOutput {
		// Run interactive input and take over terminal
		input := c.Prompt().Input()

		c.History.Append([]string{input})
		if c.appController.getOutputMode() == GoPromptOutput {
			data := c.table.store.FetchData(input)
			// Print raw text table
			printResultToSTDOUT(data)
		}
	}
	// Run interactive input, take over terminal and save output to lastStatement and statements
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
	completerWithHistoryAndDocs := autocomplete.CompleterWithHistoryAndDocs(c.History.Data, c.getSmartCompletion)

	// We need to disable the live prefix, in case we just submited a statement
	components.LivePrefixState.IsEnabled = false

	return prompt.New(
		components.Executor,
		completerWithHistoryAndDocs,
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
				c.appController.toggleOutputMode()
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

func NewInputController(history History, t *TableController, a *ApplicationController) (c InputController) {
	// Initialization
	c.History = history
	c.smartCompletion = true
	c.table = t
	c.appController = a
	return c
}
