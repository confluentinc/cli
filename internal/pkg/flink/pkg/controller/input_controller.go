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
	"github.com/confluentinc/flink-sql-client/components"
	"github.com/confluentinc/flink-sql-client/lexer"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/confluentinc/go-prompt"
	"github.com/olekukonko/tablewriter"
)

type InputControllerInterface interface {
	RunInteractiveInput()
	Prompt() *prompt.Prompt
	GetMaxCol() (int, error)
}

type InputController struct {
	History               *History
	InitialBuffer         string
	appController         ApplicationControllerInterface
	smartCompletion       bool
	reverseISearchEnabled bool
	table                 TableControllerInterface
	prompt                *prompt.Prompt
	store                 StoreInterface
}

// Actions
// This is the main function/loop for the app
func (c *InputController) RunInteractiveInput() {

	// We check for statement result and rows so we don't leave GoPrompt in case of errors
	for {
		// Run interactive input and take over terminal
		input := c.prompt.Input()

		if c.reverseISearchEnabled {
			searchResult := c.reverseISearch()
			c.reverseISearchEnabled = false
			c.setInitialBuffer(searchResult)
			continue
		}

		c.History.Append([]string{input})

		processedStatement, err := c.store.ProcessStatement(input)
		renderMsgAndStatus(processedStatement)
		if err != nil {
			fmt.Println(err.Error())
			c.isSessionValid(err)
			continue
		}

		// TODO: we need some additional logic here to decide whether we are going to show tview or just plain text
		// for now we just show tview for all non local statement
		isLocalStatement := processedStatement.Kind == configOpUse || processedStatement.Kind == configOpSet || processedStatement.Kind == configOpReset
		if !isLocalStatement {
			processedStatement, err = c.store.FetchStatementResults(*processedStatement)
			if err != nil {
				fmt.Println(err.Error())
				c.isSessionValid(err)
				continue
			}
			c.table.SetDataAndFocus(*processedStatement)
			return
		}

		printResultToSTDOUT(processedStatement.StatementResults)
	}
}

func (c *InputController) isSessionValid(err *types.StatementError) bool {
	// exit application if user needs to authenticate again
	if err != nil && err.HttpResponseCode == http.StatusUnauthorized {
		c.appController.ExitApplication()
		return false
	}
	return true
}

func (c *InputController) setInitialBuffer(s string) {
	c.InitialBuffer = s
	c.prompt = c.Prompt()
}

func renderMsgAndStatus(statementResult *types.ProcessedStatement) {
	if statementResult == nil {
		return
	}
	if statementResult.StatusDetail != "" {
		fmt.Println(statementResult.StatusDetail)
	} else {
		fmt.Println("Statement successfully submitted. No details returned from server.")
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

func printResultToSTDOUT(statementResults []types.StatementResultColumn) {
	if len(statementResults) == 0 {
		return
	}

	//build matrix
	header := make([]string, len(statementResults))
	formattedResults := make([][]string, len(statementResults[0].Fields))
	for i := range formattedResults {
		formattedResults[i] = make([]string, len(statementResults))
	}
	//fill matrix
	for colIdx, column := range statementResults {
		header[colIdx] = column.Name
		for rowIdx, field := range column.Fields {
			formattedResults[rowIdx][colIdx] = field.Format(nil)
		}
	}
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetHeader(header)
	rawTable.AppendBulk(formattedResults)
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

	return prompt.New(
		nil,
		completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory(c.History.Data),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionSetExitCheckerOnInput(func(input string, breakline bool) bool {
			if (components.IsInputClosingSelect(input) && breakline) || c.reverseISearchEnabled {
				return true
			}
			return false
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
		prompt.OptionInitialBufferText(c.InitialBuffer),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSetLexer(lexer.Lexer),
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

func reverseISearchLivePrefix(livePrefixState *ReverseISearchLivePrefixState) func() (prefix string, useLivePrefix bool) {
	return func() (prefix string, useLivePrefix bool) {
		return livePrefixState.LivePrefix, livePrefixState.IsEnable
	}
}

func (c *InputController) reverseISearch() string {

	writer := prompt.NewStdoutWriter()

	livePrefixState := &ReverseISearchLivePrefixState{
		LivePrefix: reverseISearch,
		IsEnable:   true,
	}

	searchState := &SearchState{
		index:        len(c.History.Data),
		currentMatch: "",
	}

	in := prompt.New(
		func(s string) {},
		reverseISearchCompleter(c.History.Data, writer, searchState, livePrefixState),
		prompt.OptionSetExitCheckerOnInput(func(input string, lineBreak bool) bool {
			return !c.reverseISearchEnabled
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn:  c.exitFromSearch,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlM,
			Fn:  c.exitFromSearch,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn:  c.exitFromSearch,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlR,
			Fn:  c.exitFromSearch,
		}),
		prompt.OptionWriter(writer),
		prompt.OptionTitle("bck-i-search"),
		prompt.OptionLivePrefix(reverseISearchLivePrefix(livePrefixState)),
		prompt.OptionHistory(c.History.Data),
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionSetStatementTerminator(func(lastKeyStroke prompt.Key, buffer *prompt.Buffer) bool {
			if lastKeyStroke == prompt.ControlM {
				livePrefixState.LivePrefix = ""
				return true
			}
			return false
		}),
	)
	in.Run()
	return searchState.currentMatch
}

func (c *InputController) exitFromSearch(buffer *prompt.Buffer) {
	buffer.DeleteBeforeCursor(9999)
	c.reverseISearchEnabled = false
}

// This function fetches the current max column width for the terminal
// In other words, the amount of characters that can be displayed in one line
func (c *InputController) GetMaxCol() (int, error) {
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

func NewInputController(t TableControllerInterface, a ApplicationControllerInterface, store StoreInterface, history *History) (c InputControllerInterface) {
	inputController := &InputController{
		History:         history,
		InitialBuffer:   "",
		table:           t,
		store:           store,
		appController:   a,
		smartCompletion: true,
	}
	inputController.prompt = inputController.Prompt()
	components.PrintWelcomeHeader()

	return inputController
}
