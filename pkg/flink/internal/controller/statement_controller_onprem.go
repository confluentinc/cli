package controller

import (
	"context"
	"net/http"
	"time"

	fColor "github.com/fatih/color"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v4/pkg/color"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type StatementControllerOnPrem struct {
	applicationController types.ApplicationControllerInterface
	store                 types.StoreInterfaceOnPrem
	consoleParser         prompt.ConsoleParser
	createdStatementName  string
}

func NewStatementControllerOnPrem(applicationController types.ApplicationControllerInterface, store types.StoreInterfaceOnPrem, consoleParser prompt.ConsoleParser) types.StatementControllerInterfaceOnPrem {
	return &StatementControllerOnPrem{
		applicationController: applicationController,
		store:                 store,
		consoleParser:         consoleParser,
	}
}

func (c *StatementControllerOnPrem) ExecuteStatement(statementToExecute string) (*types.ProcessedStatement, *types.StatementError) {
	processedStatement, err := c.store.ProcessStatement(statementToExecute)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}

	if processedStatement.IsDryRunStatement() {
		processedStatement.PrintOutputDryRunStatement()
		return processedStatement, nil
	}

	c.createdStatementName = processedStatement.StatementName
	processedStatement.PrintStatusMessage()

	processedStatement, err = c.waitForStatementToBeReadyOrError(*processedStatement)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}

	processedStatement, err = c.waitForStatementToBeInTerminalStateOrError(*processedStatement)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}
	processedStatement.PrintStatementDoneStatus()

	return processedStatement, nil
}

func (c *StatementControllerOnPrem) handleStatementError(err types.StatementError) {
	utils.OutputErr(err.Error())
	if err.StatusCode == http.StatusUnauthorized {
		c.applicationController.ExitApplication()
	}
}

func (c *StatementControllerOnPrem) waitForStatementToBeReadyOrError(processedStatement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	ctx, cancelWaitPendingStatement := context.WithCancel(context.Background())
	defer cancelWaitPendingStatement()

	go utils.WithPanicRecovery(func() {
		c.listenForUserInputEvent(ctx, c.userInputIsOneOf(isCancelEvent), cancelWaitPendingStatement)
	})()

	readyStatement, err := c.store.WaitPendingStatement(ctx, processedStatement)
	if err != nil {
		return nil, err
	}
	return readyStatement, nil
}

func (c *StatementControllerOnPrem) listenForUserInputEvent(ctx context.Context, userInputEvent func() bool, cancelFunc func()) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if userInputEvent() {
				cancelFunc()
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *StatementControllerOnPrem) waitForStatementToBeInTerminalStateOrError(processedStatement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	readyStatementWithResults, err := c.store.FetchStatementResults(processedStatement)
	if err != nil {
		return nil, err
	}

	if readyStatementWithResults.IsTerminalState() {
		return readyStatementWithResults, nil
	}

	ctx, cancelWaitForTerminalStatementState := context.WithCancel(context.Background())
	defer cancelWaitForTerminalStatementState()

	go utils.WithPanicRecovery(func() {
		c.listenForUserInputEvent(ctx, c.userInputIsOneOf(isDetachEvent, isCancelEvent), cancelWaitForTerminalStatementState)
	})()

	output.Printf(false, "Statement phase is %s.\n", readyStatementWithResults.Status)
	col := fColor.New(color.AccentColor)
	output.Printf(false, "Listening for execution errors. %s.\n", col.Sprint("Press Enter to detach"))
	terminalStatement, err := c.store.WaitForTerminalStatementState(ctx, *readyStatementWithResults)
	if err != nil {
		return nil, err
	}
	return terminalStatement, nil
}

func (c *StatementControllerOnPrem) userInputIsOneOf(keyEvents ...func(key prompt.Key) bool) func() bool {
	return func() bool {
		if b, err := c.consoleParser.Read(); err == nil && len(b) > 0 {
			pressedKey := prompt.Key(b[0])

			for _, isKeyEvent := range keyEvents {
				if isKeyEvent(pressedKey) {
					return true
				}
			}
		}
		return false
	}
}

func (c *StatementControllerOnPrem) CleanupStatement() {
	c.store.DeleteStatement(c.createdStatementName)
}
