package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/reverseisearch"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/autocomplete"
	lexer "github.com/confluentinc/cli/internal/pkg/flink/internal/highlighting"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
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
	History               *history.History
	InitialBuffer         string
	appController         ApplicationControllerInterface
	smartCompletion       bool
	reverseISearchEnabled bool
	table                 TableControllerInterface
	prompt                *prompt.Prompt
	store                 store.StoreInterface
	authenticated         func() error
	appOptions            *types.ApplicationOptions
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
	CANCELLED ResultsFetchState = "CANCELLED"
	COMPLETED ResultsFetchState = "COMPLETED"
)

// Actions
// This is the main function/loop for the app
func (c *InputController) RunInteractiveInput() {

	// We check for statement result and rows so we don't leave GoPrompt in case of errors
	for {
		// We save and restore the stdinState to avoid any terminal settings/shortcut bindings/Signals that can be caught and handled
		// to be unconfigured by GoPrompt. This change is smart for multiple purposes but
		// it was first introduced due to a bug where CtrlC stopped working after executing GoPrompt.
		// Ref: https://confluentinc.atlassian.net/jira/software/projects/KFS/boards/691?selectedIssue=KFS-808
		stdinState := getStdin()
		// Run interactive input and take over terminal
		input := c.prompt.Input()
		restoreStdin(stdinState)

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

		in := prompt.NewStandardInputParser()
		in.Setup()
		cancelListenToUserInput := c.listenToUserInput(in, func() {
			go c.store.DeleteStatement(processedStatement.StatementName)
			cancelWaitPendingStatement()
		})

		processedStatement, err = c.store.WaitPendingStatement(ctx, *processedStatement)
		if err != nil {
			in.TearDown()
			cancelListenToUserInput()
			fmt.Println(err.Error())
			c.isSessionValid(err)
			continue
		}

		processedStatement, err = c.store.FetchStatementResults(*processedStatement)
		in.TearDown()
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
					statementResults, _ := results.ConvertToInternalResults(*mockExample.StatementResults.Results.Data, mockExample.ResultSchema)
					processedStatement.StatementResults = statementResults
					processedStatement.PageToken = ""
				}
			}
			c.table.Init(*processedStatement)
			return
		}

		c.printResultToSTDOUT(processedStatement.StatementResults)
		// We already printed the results using plain text and will delete the statement. When using TView this will happen upon leaving the interactive view.
		// TODO - this is currently used only to save system resources, To be removed once the API Server becomes scalable.
		// We want to maintain a "completed" statement in the backend
		if !processedStatement.IsLocalStatement && processedStatement.Status != types.RUNNING {
			go c.store.DeleteStatement(processedStatement.StatementName)
		}
	}
}

func (c *InputController) listenToUserInput(in *prompt.PosixParser, cancelFunc context.CancelFunc) context.CancelFunc {
	ctx, cancelListenToUserInput := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if b, err := in.Read(); err == nil && len(b) > 0 {
					pressedKey := prompt.Key(b[0])

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
			time.Sleep(10 * time.Millisecond)
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

	if statementResult.IsLocalStatement {
		if statementResult.Status != "FAILED" {
			fmt.Println("Statement successfully submitted.\n ")
		} else {
			fmt.Println("Error: Couldn't process statement. Please check your statement and try again.")
		}
	} else {

		if statementResult.StatementName != "" {
			fmt.Println("Statement ID: " + statementResult.StatementName)
		}
		if statementResult.Status != "FAILED" {
			fmt.Println("Statement successfully submitted. ")
			fmt.Println("Fetching results...\n ")
		} else {
			fmt.Println("Error: Statement submission failed. There could a problem with the server right now. Check your statement and try again.")
		}
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

	components.PrintOutputModeState(c.appController.GetOutputMode() == types.TViewOutput, maxCol)
}

func (c *InputController) printResultToSTDOUT(statementResults *types.StatementResults) {
	if statementResults == nil || len(statementResults.Headers) == 0 || len(statementResults.Rows) == 0 {
		fmt.Println("\nThe server returned empty rows for this statement.")
		return
	}

	windowSize, err := c.GetMaxCol()
	if err != nil {
		// set a default size on error
		windowSize = 100
	}

	fixedPadding := 4                                          // table border left and right
	variablePadding := (len(statementResults.Headers) - 1) * 3 // column separator
	totalAvailableChars := windowSize - fixedPadding - variablePadding
	charsPerColumn := totalAvailableChars / len(statementResults.Headers) //distribute chars evenly
	formatterOptions := &types.FormatterOptions{MaxCharCountToDisplay: charsPerColumn}

	var formattedResults [][]string
	for _, row := range statementResults.Rows {
		var formattedRow []string
		for _, field := range row.Fields {
			formattedRow = append(formattedRow, field.Format(formatterOptions))
		}
		formattedResults = append(formattedResults, formattedRow)
	}
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetAutoFormatHeaders(false)
	rawTable.SetHeader(statementResults.Headers)
	rawTable.AppendBulk(formattedResults)
	rawTable.Render() // Send output
}

func (c *InputController) Prompt() *prompt.Prompt {
	completer := autocomplete.NewCompleterBuilder(c.getSmartCompletion).
		AddCompleter(autocomplete.ExamplesCompleter).
		AddCompleter(autocomplete.SetCompleter).
		AddCompleter(autocomplete.ShowCompleter).
		AddCompleter(autocomplete.GenerateHistoryCompleter(c.History.Data)).
		AddCompleter(autocomplete.GenerateDocsCompleter()).
		BuildCompleter()

	return prompt.New(
		nil,
		completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory(c.History.Data),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionSetExitCheckerOnInput(func(input string, breakline bool) bool {
			if c.reverseISearchEnabled || c.shouldExit {
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

func reverseISearchLivePrefix(livePrefixState *reverseisearch.LivePrefixState) func() (prefix string, useLivePrefix bool) {
	return func() (prefix string, useLivePrefix bool) {
		return livePrefixState.LivePrefix, livePrefixState.IsEnable
	}
}

func (c *InputController) reverseISearch() string {

	writer := prompt.NewStdoutWriter()

	livePrefixState := &reverseisearch.LivePrefixState{
		LivePrefix: reverseisearch.BckISearch,
		IsEnable:   true,
	}

	searchState := &reverseisearch.SearchState{
		CurrentIndex: len(c.History.Data) - 1,
		CurrentMatch: "",
	}

	in := prompt.New(
		func(s string) {},
		reverseisearch.SearchCompleter(c.History.Data, writer, searchState, livePrefixState),
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
			Fn:  reverseisearch.NextResult(writer, c.History.Data, searchState, livePrefixState),
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
	return searchState.CurrentMatch
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

func NewInputController(t TableControllerInterface, a ApplicationControllerInterface, store store.StoreInterface, authenticated func() error, history *history.History, appOptions *types.ApplicationOptions) (c InputControllerInterface) {
	inputController := &InputController{
		History:         history,
		InitialBuffer:   "",
		table:           t,
		store:           store,
		appController:   a,
		smartCompletion: false,
		authenticated:   authenticated,
		appOptions:      appOptions,
		shouldExit:      false,
	}
	inputController.prompt = inputController.Prompt()
	components.PrintWelcomeHeader()

	return inputController
}
