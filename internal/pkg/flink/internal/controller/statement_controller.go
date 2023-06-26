package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type StatementController struct {
	applicationController types.ApplicationControllerInterface
	store                 store.StoreInterface
	consoleParser         prompt.ConsoleParser
}

func NewStatementController(applicationController types.ApplicationControllerInterface, store store.StoreInterface, consoleParser prompt.ConsoleParser) types.StatementControllerInterface {
	return &StatementController{
		applicationController: applicationController,
		store:                 store,
		consoleParser:         consoleParser,
	}
}

func (c *StatementController) ExecuteStatement(statementToExecute string) (*types.ProcessedStatement, *types.StatementError) {
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

func (c *StatementController) handleStatementError(err types.StatementError) {
	utils.OutputErr(err.Error())
	if !isSessionValid(err) {
		c.applicationController.ExitApplication()
	}
}

func isSessionValid(err types.StatementError) bool {
	return err.HttpResponseCode != http.StatusUnauthorized
}

func renderMsgAndStatus(processedStatement types.ProcessedStatement) {
	if processedStatement.IsLocalStatement {
		renderMsgAndStatusOfLocalStatement(processedStatement)
	} else {
		renderMsgAndStatusOfNonLocalStatement(processedStatement)
	}
}

func renderMsgAndStatusOfLocalStatement(processedStatement types.ProcessedStatement) {
	if processedStatement.Status == "FAILED" {
		err := types.StatementError{Message: "couldn't process statement, please check your statement and try again"}
		utils.OutputErr(err.Error())
	} else {
		utils.OutputInfo("Statement successfully submitted.")
	}
}

func renderMsgAndStatusOfNonLocalStatement(processedStatement types.ProcessedStatement) {
	if processedStatement.StatementName != "" {
		utils.OutputInfof("Statement name: %s\n", processedStatement.StatementName)
	}
	if processedStatement.Status == "FAILED" {
		err := types.StatementError{Message: "statement submission failed"}
		utils.OutputErr(err.Error())
	} else {
		utils.OutputInfo("Statement successfully submitted.")
		utils.OutputInfo("Fetching results...")
	}
}

func (c *StatementController) waitForStatementToBeReadyOrError(processedStatement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	ctx, cancelWaitPendingStatement := context.WithCancel(context.Background())
	defer cancelWaitPendingStatement()

	statementName := processedStatement.StatementName
	go c.listenForUserCancelEvent(ctx, func() {
		cancelWaitPendingStatement()
		c.store.DeleteStatement(statementName)
	})

	readyStatement, err := c.store.WaitPendingStatement(ctx, processedStatement)
	if err != nil {
		return nil, err
	}
	return readyStatement, nil
}

func (c *StatementController) listenForUserCancelEvent(ctx context.Context, cancelFunc func()) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if c.isCancelEvent() {
				cancelFunc()
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *StatementController) isCancelEvent() bool {
	if b, err := c.consoleParser.Read(); err == nil && len(b) > 0 {
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
			return true
		}
	}
	return false
}
