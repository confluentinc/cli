package controller

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	testUtils "github.com/confluentinc/cli/v4/pkg/flink/test"
	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
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
	processedStatementError := types.StatementError{Message: "unauthorized", StatusCode: http.StatusUnauthorized}
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
		Principal:     "sa-123",
	}
	completedStatement := processedStatement
	completedStatement.Status = types.COMPLETED
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&completedStatement, nil)
	s.store.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		_, _ = s.statementController.ExecuteStatement(statementToExecute)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementPrintsNoWarningWhenUserIdentityIsUsed() {
	statementToExecute := "insert into table values (1,2);"
	processedStatement := types.ProcessedStatement{
		Statement:     statementToExecute,
		StatementName: "test-statement",
		StatusDetail:  "status detail message",
		Status:        types.PENDING,
		Principal:     "u-123",
	}
	completedStatement := processedStatement
	completedStatement.Status = types.COMPLETED
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&completedStatement, nil)
	s.store.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		_, _ = s.statementController.ExecuteStatement(statementToExecute)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementPrintsNoWarningForLocalStatements() {
	statementToExecute := "insert into table values (1,2);"
	processedStatement := types.ProcessedStatement{
		Statement:        statementToExecute,
		StatementName:    "test-statement",
		StatusDetail:     "status detail message",
		Status:           types.PENDING,
		Principal:        "u-123",
		IsLocalStatement: true,
	}
	completedStatement := processedStatement
	completedStatement.Status = types.COMPLETED
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&completedStatement, nil)
	s.store.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		_, _ = s.statementController.ExecuteStatement(statementToExecute)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementPrintsNoWarningForStatementsOtherThanInsertOrStatementSet() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{
		Statement:     statementToExecute,
		StatementName: "test-statement",
		StatusDetail:  "status detail message",
		Status:        types.PENDING,
		Principal:     "u-123",
	}
	completedStatement := processedStatement
	completedStatement.Status = types.COMPLETED
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&completedStatement, nil)
	s.store.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		_, _ = s.statementController.ExecuteStatement(statementToExecute)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWaitsForCompletedState() {
	statementToExecute := "insert into users values ('test');"
	processedStatement := types.ProcessedStatement{Status: types.PENDING, Principal: "sa-123"}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	completedStatement := types.ProcessedStatement{Status: types.COMPLETED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).Return(&completedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &completedStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWaitsForFailedState() {
	statementToExecute := "insert into users values ('test');"
	processedStatement := types.ProcessedStatement{Status: types.PENDING, Principal: "sa-123"}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	failedStatement := types.ProcessedStatement{Status: types.FAILED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).Return(&failedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &failedStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWaitsForNonEmptyPageToken() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{Status: types.PENDING, Principal: "sa-123"}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING, PageToken: "not-empty"}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &runningStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementReturnsWhenUserDetaches() {
	statementToExecute := "insert into users values ('test');"
	processedStatement := types.ProcessedStatement{Status: types.PENDING, Principal: "sa-123"}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.consoleParser.EXPECT().Read().Return([]byte{byte(prompt.ControlM)}, nil).AnyTimes()
	var waitForTerminalStateCtx context.Context
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).DoAndReturn(
		func(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
			waitForTerminalStateCtx = ctx
			time.Sleep(time.Second)
			return &runningStatement, nil
		})

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Error(s.T(), waitForTerminalStateCtx.Err())
		require.Equal(s.T(), &runningStatement, returnedStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *StatementControllerTestSuite) TestExecuteStatementWithWarning() {
	statementToExecute := "insert into users values ('test');"
	warning := "[Warning] The primary key does not match the upsert key derived from the query. If the primary key and upsert key don't match, the system needs to add a state-intensive operation for correction. Please revisit the query (upsert key: [customer_name]) or the table declaration for `default`.`dmvk`.`o` (primary key: [total_orders])"
	processedStatement := types.ProcessedStatement{Status: types.PENDING, Principal: "sa-123", StatusDetail: warning}
	runningStatement := types.ProcessedStatement{Status: types.RUNNING, StatusDetail: warning}
	completedStatement := types.ProcessedStatement{Status: types.COMPLETED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&runningStatement, nil)
	s.store.EXPECT().FetchStatementResults(runningStatement).Return(&runningStatement, nil)
	s.store.EXPECT().WaitForTerminalStatementState(gomock.Any(), runningStatement).Return(&completedStatement, nil)

	stdout := testUtils.RunAndCaptureSTDOUT(s.T(), func() {
		returnedStatement, err := s.statementController.ExecuteStatement(statementToExecute)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &completedStatement, returnedStatement)
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
		},
		{
			name:      "local non-failed statement",
			statement: types.ProcessedStatement{IsLocalStatement: true, Status: types.RUNNING},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := testUtils.RunAndCaptureSTDOUT(s.T(), test.statement.PrintStatusMessage)
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
		},
		{
			name:      "statement without name",
			statement: types.ProcessedStatement{Status: types.FAILED},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := testUtils.RunAndCaptureSTDOUT(s.T(), test.statement.PrintStatusMessage)
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
		},
		{
			name:      "statement without name",
			statement: types.ProcessedStatement{Status: types.RUNNING},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := testUtils.RunAndCaptureSTDOUT(s.T(), test.statement.PrintStatusMessage)
			cupaloy.SnapshotT(t, actual)
		})
	}
}

func (s *StatementControllerTestSuite) TestOutputOfDryRunStatements() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      string
	}{
		{
			name:      "failed dry run",
			statement: types.ProcessedStatement{StatementName: "failed-test-statement", Status: types.FAILED, StatusDetail: "Parse Error"},
		},
		{
			name:      "successful dry run",
			statement: types.ProcessedStatement{StatementName: "successful-test-statement", Status: types.COMPLETED, StatusDetail: "No errors"},
		},
		{
			name:      "unexpected status dry run",
			statement: types.ProcessedStatement{StatementName: "unexpected-test-statement", Status: types.RUNNING, StatusDetail: "unexpected status"},
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := testUtils.RunAndCaptureSTDOUT(s.T(), test.statement.PrintOutputDryRunStatement)
			cupaloy.SnapshotT(t, actual)
		})
	}
}

func (s *StatementControllerTestSuite) TestIsDryRunStatements() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		expected  bool
	}{
		{
			name: "is dry run statement",
			statement: types.ProcessedStatement{Properties: map[string]string{
				config.KeyDryRun: "true",
			}},
			expected: true,
		},
		{
			name: "dry run is false",
			statement: types.ProcessedStatement{Properties: map[string]string{
				config.KeyDryRun: "false",
			}},
			expected: false,
		},
		{
			name:      "dry run not set",
			statement: types.ProcessedStatement{},
			expected:  false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := test.statement.IsDryRunStatement()
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestIsCancelEvent(t *testing.T) {
	tests := []struct {
		name string
		key  prompt.Key
		want bool
	}{
		{
			name: "ControlC",
			key:  prompt.ControlC,
			want: true,
		},
		{
			name: "ControlD",
			key:  prompt.ControlD,
			want: true,
		},
		{
			name: "ControlQ",
			key:  prompt.ControlQ,
			want: true,
		},
		{
			name: "Escape",
			key:  prompt.ControlSpace,
			want: true,
		},
		{
			name: "Other",
			key:  prompt.ShiftDown, // Just an example of a key that is not in the switch cases
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isCancelEvent(test.key)
			require.Equal(t, test.want, got)
		})
	}
}

func TestIsDetachEvent(t *testing.T) {
	tests := []struct {
		name string
		key  prompt.Key
		want bool
	}{
		{
			name: "ControlM",
			key:  prompt.ControlM,
			want: true,
		},
		{
			name: "Enter",
			key:  prompt.Enter,
			want: true,
		},
		{
			name: "Other",
			key:  prompt.ShiftUp, // Just an example of a key that is not in the switch cases
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isDetachEvent(test.key)
			require.Equal(t, test.want, got)
		})
	}
}

func (s *StatementControllerTestSuite) TestCleanupStatement() {
	statementToExecute := "select 1;"
	processedStatement := types.ProcessedStatement{StatementName: "test-statement", Status: types.COMPLETED}
	s.store.EXPECT().ProcessStatement(statementToExecute).Return(&processedStatement, nil)
	s.consoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.store.EXPECT().WaitPendingStatement(gomock.Any(), processedStatement).Return(&processedStatement, nil)
	s.store.EXPECT().FetchStatementResults(processedStatement).Return(&processedStatement, nil)
	_, err := s.statementController.ExecuteStatement(statementToExecute)
	require.Nil(s.T(), err)

	require.Equal(s.T(), processedStatement.StatementName, s.statementController.(*StatementController).createdStatementName)
	s.store.EXPECT().DeleteStatement(processedStatement.StatementName)
	s.statementController.CleanupStatement()
}
