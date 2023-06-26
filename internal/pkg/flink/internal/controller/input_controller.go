package controller

import (
	"context"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/autocomplete"
	lexer "github.com/confluentinc/cli/internal/pkg/flink/internal/highlighting"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/reverseisearch"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type InputController struct {
	History               *history.History
	InitialBuffer         string
	appController         types.ApplicationControllerInterface
	smartCompletion       bool
	reverseISearchEnabled bool
	table                 types.TableControllerInterface
	prompt                prompt.IPrompt
	store                 store.StoreInterface
	authenticated         func() error
	appOptions            *types.ApplicationOptions
	shouldExit            bool
	stdin                 *term.State
	consoleParser         prompt.ConsoleParser
	reverseISearch        reverseisearch.ReverseISearch
}

func shouldUseTView(statement types.ProcessedStatement) bool {
	// only use view for non-local statements, that have more than one row and more than one column
	if statement.IsLocalStatement {
		return false
	}
	if statement.PageToken != "" {
		return true
	}
	return len(statement.StatementResults.GetHeaders()) > 1 && len(statement.StatementResults.GetRows()) > 1
}

// This is the main function/loop for the app
func (c *InputController) RunInteractiveInput() {
	for c.shouldRunMainLoop() {
		userInput := c.getUserInput()
		if c.isSpecialInput(userInput) {
			continue
		}
		c.History.Append([]string{userInput})

		executedStatement, err := c.executeStatement(userInput)
		if err != nil {
			continue
		}

		if c.handleStatementResults(executedStatement) {
			return
		}
	}
}

func (c *InputController) shouldRunMainLoop() bool {
	if authErr := c.authenticated(); authErr != nil {
		outputErrf("Error: %v\n", authErr)
		c.appController.ExitApplication()
		return false
	}
	return !c.shouldExit
}

func (c *InputController) getUserInput() string {
	// if the initial buffer is not empty, we insert the text and reset the InitialBuffer
	if c.InitialBuffer != "" {
		c.prompt.Buffer().InsertText(c.InitialBuffer, false, true)
		c.InitialBuffer = ""
	}
	return c.prompt.Input()
}

func (c *InputController) isSpecialInput(userInput string) bool {
	if c.reverseISearchEnabled {
		searchResult := c.reverseISearch.ReverseISearch(c.History.Data)
		c.reverseISearchEnabled = false
		c.InitialBuffer = searchResult
		return true
	}

	if c.shouldExit || userInput == "" {
		c.appController.ExitApplication()
		return true
	}

	return false
}

func (c *InputController) executeStatement(statementToExecute string) (*types.ProcessedStatement, *types.StatementError) {
	processedStatement, err := c.store.ProcessStatement(statementToExecute)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}
	renderMsgAndStatus(*processedStatement)

	processedStatement, err = c.waitForStatementToBeReadyOrError(*processedStatement)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}
	processedStatement.PrintStatusDetail()

	processedStatement, err = c.store.FetchStatementResults(*processedStatement)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}

	return processedStatement, nil
}

func (c *InputController) handleStatementError(err types.StatementError) {
	outputErr(err.Error())
	if !c.isSessionValid(err) {
		c.appController.ExitApplication()
	}
}

func (c *InputController) waitForStatementToBeReadyOrError(processedStatement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	ctx, cancelWaitPendingStatement := context.WithCancel(context.Background())
	statementName := processedStatement.StatementName
	cancelListenForUserCancelEvent := c.listenForUserCancelEvent(c.consoleParser, func() {
		go c.store.DeleteStatement(statementName)
		cancelWaitPendingStatement()
	})
	defer cancelListenForUserCancelEvent()

	readyStatement, err := c.store.WaitPendingStatement(ctx, processedStatement)
	if err != nil {
		return nil, err
	}
	return readyStatement, nil
}

func (c *InputController) handleStatementResults(processedStatement *types.ProcessedStatement) bool {
	// decide if we want to display results using TView or just a plain table
	if shouldUseTView(*processedStatement) {
		c.table.Init(*processedStatement)
		return true
	}

	c.printResultToSTDOUT(processedStatement.StatementResults)
	// This was used to delete statements after their execution to save system resources, which should not be
	// an issue anymore. We don't want to remove it completely just yet, but will disable it by default for now.
	// TODO: remove this completely once we are sure we won't need it in the future
	if config.ShouldCleanupStatements && !processedStatement.IsLocalStatement && processedStatement.Status != types.RUNNING {
		go c.store.DeleteStatement(processedStatement.StatementName)
	}
	return false
}

func (c *InputController) listenForUserCancelEvent(in prompt.ConsoleParser, cancelFunc context.CancelFunc) context.CancelFunc {
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

func (c *InputController) isSessionValid(err types.StatementError) bool {
	// exit application if user needs to authenticate again
	if err.HttpResponseCode == http.StatusUnauthorized {
		return false
	}
	return true
}

func renderMsgAndStatus(processedStatement types.ProcessedStatement) {
	if processedStatement.IsLocalStatement {
		if processedStatement.Status == "FAILED" {
			err := types.StatementError{Message: "couldn't process statement, please check your statement and try again"}
			outputErr(err.Error())
		} else {
			outputInfo("Statement successfully submitted.")
		}
	} else {
		if processedStatement.StatementName != "" {
			outputInfof("Statement name: %s\n", processedStatement.StatementName)
		}
		if processedStatement.Status == "FAILED" {
			err := types.StatementError{Message: "statement submission failed"}
			outputErr(err.Error())
		} else {
			outputInfo("Statement successfully submitted.")
			outputInfo("Fetching results...")
		}
		processedStatement.PrintStatusDetail()
	}
}

func (c *InputController) toggleSmartCompletion() {
	c.smartCompletion = !c.smartCompletion

	maxCol, err := c.GetMaxCol()
	if err != nil {
		log.CliLogger.Error(err)
		return
	}

	components.PrintSmartCompletionState(c.getSmartCompletion(), maxCol)
}

func (c *InputController) toggleOutputMode() {
	c.appController.ToggleOutputMode()

	maxCol, err := c.GetMaxCol()
	if err != nil {
		log.CliLogger.Error(err)
		return
	}

	components.PrintOutputModeState(c.appController.GetOutputMode() == types.TViewOutput, maxCol)
}

func (c *InputController) printResultToSTDOUT(statementResults *types.StatementResults) {
	if statementResults == nil || len(statementResults.Headers) == 0 || len(statementResults.Rows) == 0 {
		outputWarn("\nThe server returned empty rows for this statement.")
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

	materializedStatementResults := types.NewMaterializedStatementResults(statementResults.GetHeaders(), maxResultsCapacity)
	materializedStatementResults.Append(statementResults.GetRows()...)
	columnWidths := materializedStatementResults.GetMaxWidthPerColum()
	columnWidths = results.GetTruncatedColumnWidths(columnWidths, totalAvailableChars)

	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetAutoFormatHeaders(false)
	rawTable.SetHeader(statementResults.Headers)
	// add actual row data
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		formattedRow := make([]string, len(row.Fields))
		for colIdx, field := range row.Fields {
			formattedRow[colIdx] = results.TruncateString(field.ToString(), columnWidths[colIdx])
		}
		rawTable.Append(formattedRow)
	})
	rawTable.Render() // Send output
}

func (c *InputController) Prompt() prompt.IPrompt {
	completer := autocomplete.NewCompleterBuilder(c.getSmartCompletion).
		AddCompleter(autocomplete.ExamplesCompleter).
		AddCompleter(autocomplete.SetCompleter).
		AddCompleter(autocomplete.ShowCompleter).
		AddCompleter(autocomplete.GenerateHistoryCompleter(c.History.Data)).
		// AddCompleter(autocomplete.GenerateDocsCompleter()).
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

// Getters
func (c *InputController) getSmartCompletion() bool {
	return c.smartCompletion
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

func (c *InputController) tearDown() {
	tearDownConsoleParser(c.consoleParser)
	restoreStdin(c.stdin)
}

func NewInputController(t types.TableControllerInterface, a types.ApplicationControllerInterface, store store.StoreInterface, authenticated func() error, history *history.History, appOptions *types.ApplicationOptions) types.InputControllerInterface {
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
		stdin:           getStdin(),
		consoleParser:   getConsoleParser(),
		reverseISearch:  reverseisearch.NewReverseISearch(),
	}
	a.AddCleanupFunction(inputController.tearDown)
	inputController.prompt = inputController.Prompt()
	components.PrintWelcomeHeader()

	return inputController
}
