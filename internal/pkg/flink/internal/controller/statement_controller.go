package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type StatementController struct {
	applicationController types.ApplicationControllerInterface
	store                 types.StoreInterface
	consoleParser         prompt.ConsoleParser
}

func NewStatementController(applicationController types.ApplicationControllerInterface, store types.StoreInterface, consoleParser prompt.ConsoleParser) types.StatementControllerInterface {
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
	processedStatement.PrintStatusMessage()

	processedStatement, err = c.waitForStatementToBeReadyOrError(*processedStatement)
	if err != nil {
		c.handleStatementError(*err)
		return nil, err
	}
	processedStatement.PrintStatementDoneStatus()

	return processedStatement, nil
}

func (c *StatementController) handleStatementError(err types.StatementError) {
	utils.OutputErr(err.Error())
	if err.HttpResponseCode == http.StatusUnauthorized {
		c.applicationController.ExitApplication()
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
