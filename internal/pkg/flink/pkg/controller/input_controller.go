package controller

import (
	"bufio"
	"context"
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
	"github.com/confluentinc/flink-sql-client/pkg/results"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/confluentinc/flink-sql-client/test/generators"
	"github.com/confluentinc/go-prompt"
	"github.com/olekukonko/tablewriter"
	"pgregory.net/rapid"
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
	authenticated         func() error
	appOptions            *ApplicationOptions
	shouldExit            bool
}

func shouldUseTView(statement types.ProcessedStatement) bool {
	// only use view for non-local statements, that have more than one row and more than one column
	if statement.IsLocalStatement {
		return false
	}
	if statement.PageToken != "" {
		return true
	}
	return len(statement.StatementResults.Headers) > 1 && len(statement.StatementResults.Rows) > 1
}

type ResultsFetchState string

const (
	PENDING   ResultsFetchState = "PENDING"
	STARTED   ResultsFetchState = "STARTED"
	CANCELED  ResultsFetchState = "CANCELED"
	COMPLETED ResultsFetchState = "COMPLETED"
)

// Actions
// This is the main function/loop for the app
func (c *InputController) RunInteractiveInput() {

	// We check for statement result and rows so we don't leave GoPrompt in case of errors
	for {
		// Run interactive input and take over terminal
		input := c.prompt.Input()

		if c.shouldExit {
			c.appController.ExitApplication()
		}

		// Upon receiving user input, we check if user is authenticated and possibly a refresh the CCloud SSO token
		if authErr := c.authenticated(); authErr != nil {
			fmt.Println(authErr.Error())
			c.appController.ExitApplication()
			continue
		}

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

		// Wait for results to be there or for the user to cancel the query
		ctx, cancelWaitPendingStatement := context.WithCancel(context.Background())
		cancelListenToUserInput := c.listenToUserInput(func() {
			c.store.DeleteStatement(processedStatement.StatementName)
			cancelWaitPendingStatement()
		})

		processedStatement, err = c.store.WaitPendingStatement(ctx, *processedStatement)
		if err != nil {
			cancelListenToUserInput()
			fmt.Println(err.Error())
			c.isSessionValid(err)
			continue
		}

		processedStatement, err = c.store.FetchStatementResults(*processedStatement)
		cancelListenToUserInput()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// decide if we want to display results using TView or just a plain table
		if shouldUseTView(*processedStatement) {
			demoMode := c.appOptions != nil && c.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO
			if demoMode {
				if rapid.Bool().Example() {
					mockExample := generators.MockResults(5, 2).Example()
					statementresults, _ := results.ConvertToInternalResults(*mockExample.StatementResults.Results.Data, mockExample.ResultSchema)
					processedStatement.StatementResults = statementresults
					processedStatement.PageToken = ""
				}
			}
			c.table.Init(*processedStatement)
			return
		}

		printResultToSTDOUT(processedStatement.StatementResults)
	}
}

func (c *InputController) listenToUserInput(cancelFunc context.CancelFunc) context.CancelFunc {
	ctx, cancelListenToUserInput := context.WithCancel(context.Background())
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				input, err := reader.ReadByte()
				if err != nil {
					continue
				}
				pressedKey := prompt.Key(input)

				switch pressedKey {
				case prompt.ControlC:
					fallthrough
				case prompt.ControlD:
					fallthrough
				case prompt.ControlQ:
					fallthrough
				case prompt.Escape:
					// esc
					cancelFunc()
					return
				}
			}
		}
	}()
	return cancelListenToUserInput
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

func printResultToSTDOUT(statementResults *types.StatementResults) {
	if statementResults == nil || len(statementResults.Headers) == 0 || len(statementResults.Rows) == 0 {
		fmt.Println("\nThe server returned empty rows for this statement.")
		return
	}

	var formattedResults [][]string
	for _, row := range statementResults.Rows {
		var formattedRow []string
		for _, field := range row.Fields {
			formattedRow = append(formattedRow, field.Format(nil))
		}
		formattedResults = append(formattedResults, formattedRow)
	}
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetHeader(statementResults.Headers)
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
			if (components.IsInputClosingSelect(input) && breakline) || c.reverseISearchEnabled || c.shouldExit {
				return true
			}
			return false
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

func NewInputController(t TableControllerInterface, a ApplicationControllerInterface, store StoreInterface, authenticated func() error, history *History, appOptions *ApplicationOptions) (c InputControllerInterface) {
	inputController := &InputController{
		History:         history,
		InitialBuffer:   "",
		table:           t,
		store:           store,
		appController:   a,
		smartCompletion: true,
		authenticated:   authenticated,
		appOptions:      appOptions,
		shouldExit:      false,
	}
	inputController.prompt = inputController.Prompt()
	components.PrintWelcomeHeader()

	return inputController
}
