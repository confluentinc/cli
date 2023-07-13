package controller

import (
	"context"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/test"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type StatementControllerTestSuite struct {
	suite.Suite
	statementController   types.StatementControllerInterface
	applicationController *mock.MockApplicationControllerInterface
	store                 *mock.MockStoreInterface
	consoleParser         *mock.MockConsoleParser
}

func TestStatementControllerTestSuite(t *testing.T) {
	suite.Run(t, new(StatementControllerTestSuite))
}

func (s *StatementControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.applicationController = mock.NewMockApplicationControllerInterface(ctrl)
	s.store = mock.NewMockStoreInterface(ctrl)
	s.consoleParser = mock.NewMockConsoleParser(ctrl)
	s.statementController = NewStatementController(s.applicationController, s.store, s.consoleParser)
}

func (s *StatementControllerTestSuite) TestExecuteStatementReturnsProcessStatementError() {
	statementToExecute := "select 1;"
	processStatementError := &types.StatementError{Message: "processing error"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(nil, processStatementError)

	_, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Equal(s.T(), processStatementError, err)
}

func (s *StatementControllerTestSuite) TestExecuteStatementReturnsWaitForStatementError() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{}
	waitPendingStatementError := &types.StatementError{Message: "wait error"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(nil, waitPendingStatementError)

	_, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Equal(s.T(), waitPendingStatementError, err)
}

func (s *StatementControllerTestSuite) TestExecuteStatementReturnsFetchStatementResultsError() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{}
	fetchStatementResultsError := &types.StatementError{Message: "fetch results error"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&processedStatement, nil)
	s.store.EXPECT().FetchStatementResults(processedStatement).Return(nil, fetchStatementResultsError)

	_, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Equal(s.T(), fetchStatementResultsError, err)
}

func (s *StatementControllerTestSuite) TestExecuteStatementReturnsWaitForTerminalStateError() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{}
	waitForTerminalStatementStateError := &types.StatementError{Message: "wait for terminal state error"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&processedStatement, nil)
	s.store.EXPECT().FetchStatementResults(processedStatement).Return(&processedStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), processedStatement).Return(nil, waitForTerminalStatementStateError)

	_, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Equal(s.T(), waitForTerminalStatementStateError, err)
}

func (s *StatementControllerTestSuite) TestExecuteStatement() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{Status: types.COMPLETED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&processedStatement, nil)
	s.store.EXPECT().FetchStatementResults(processedStatement).Return(&processedStatement, nil)

	returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Nil(s.T(), err)
	require.Equal(s.T(), &processedStatement, returnedStatement)
}

func (s *StatementControllerTestSuite) TestExecuteStatementExitApplicationOnUnauthorizedResponse() {
	statementToExecute := "select 1;"
	processedStatementError := types.StatementError{Message: "unauthorized", HttpResponseCode: 401}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(nil, &processedStatementError)
	s.applicationController.EXPECT().ExitApplication()

	_, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Equal(s.T(), &processedStatementError, err)
}

func (s *StatementControllerTestSuite) TestExecuteStatementCancelsAndDeletesStatementOnUserInterrupt() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{}
	waitPendingStatementError := &types.StatementError{Message: "wait error"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return([]byte{byte(prompt.ControlC)}, nil)
	var waitPendingStatementCtx context.Context
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).DoAndReturn(
		func(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
			waitPendingStatementCtx = ctx
			time.Sleep(time.Second)
			return nil, waitPendingStatementError
		})

	_, err := s.statementController.ExecuteStatement(statementToExecute)

	require.Equal(s.T(), waitPendingStatementError, err)
	require.Error(s.T(), waitPendingStatementCtx.Err())
}

func (s *StatementControllerTestSuite) TestExecuteStatementPrintsUserInfo() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{
		StatementName: "test-statement",
		StatusDetail:  "status detail message",
		Status:        types.PENDING,
	}
	completedStatement := processedStatement
	completedStatement.Status = types.COMPLETED
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&completedStatement, nil)
	s.store.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		_, _ = s.statementController.ExecuteStatement(statementToExecute)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWaitsForCompletedState() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{Status: types.PENDING}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	completedStatement := types.ProcessedStatement{Status: types.COMPLETED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).Return(&completedStatement, nil)

	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &completedStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWaitsForFailedState() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{Status: types.PENDING}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	failedStatement := types.ProcessedStatement{Status: types.FAILED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).Return(&failedStatement, nil)

	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &failedStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWaitsForNonEmptyPageToken() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{Status: types.PENDING}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	runningStatementWithNextPage := types.ProcessedStatement{Status: types.RUNNING, PageToken: "not-empty"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).Return(&runningStatementWithNextPage, nil)

	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &runningStatementWithNextPage, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementReturnsWhenUserDetaches() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{Status: types.PENDING}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return([]byte{byte(prompt.ControlM)}, nil)
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	var waitForTerminalStateCtx context.Context
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).DoAndReturn(
		func(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
			waitForTerminalStateCtx = ctx
			time.Sleep(time.Second)
			return &runningStatement, nil
		})

	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Error(s.T(), waitForTerminalStateCtx.Err())
		require.Equal(s.T(), &runningStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestRenderMsgAndStatusLocalStatements() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      string
	}{
		{
			name:      "local failed statement",
			statement: types.ProcessedStatement{IsLocalStatement: true, Status: types.FAILED},
			want:      "Error: couldn't process statement, please check your statement and try again\n",
		},
		{
			name:      "local non-failed statement",
			statement: types.ProcessedStatement{IsLocalStatement: true, Status: types.RUNNING},
			want:      "Statement successfully submitted.\n",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			actual := test.RunAndCaptureSTDOUT(s.T(), tt.statement.PrintStatusMessage)
			cupaloy.SnapshotT(t, actual)
		})
	}
}

func (s *StatementControllerTestSuite) TestRenderMsgAndStatusNonLocalFailedStatements() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      string
	}{
		{
			name:      "statement with name",
			statement: types.ProcessedStatement{StatementName: "test-statement", Status: types.FAILED},
			want:      "Statement name: test-statement\nError: statement submission failed\n",
		},
		{
			name:      "statement without name",
			statement: types.ProcessedStatement{Status: types.FAILED},
			want:      "Error: statement submission failed\n",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			actual := test.RunAndCaptureSTDOUT(s.T(), tt.statement.PrintStatusMessage)
			cupaloy.SnapshotT(t, actual)
		})
	}
}

func (s *StatementControllerTestSuite) TestRenderMsgAndStatusNonLocalNonFailedStatements() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      string
	}{
		{
			name:      "statement with name",
			statement: types.ProcessedStatement{StatementName: "test-statement", Status: types.RUNNING},
			want:      "Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n",
		},
		{
			name:      "statement without name",
			statement: types.ProcessedStatement{Status: types.RUNNING},
			want:      "Statement successfully submitted.\nFetching results...\n",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			actual := test.RunAndCaptureSTDOUT(s.T(), tt.statement.PrintStatusMessage)
			cupaloy.SnapshotT(t, actual)
		})
	}
}
