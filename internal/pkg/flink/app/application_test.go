package app

import (
	"github.com/bradleyjkemp/cupaloy"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/test"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type ApplicationTestSuite struct {
	suite.Suite
	app                         Application
	history                     *history.History
	store                       *mock.MockStoreInterface
	resultFetcher               *mock.MockResultFetcherInterface
	appController               *mock.MockApplicationControllerInterface
	inputController             *mock.MockInputControllerInterface
	statementController         *mock.MockStatementControllerInterface
	basicOutputController       *mock.MockOutputControllerInterface
	interactiveOutputController *mock.MockOutputControllerInterface
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.appController = mock.NewMockApplicationControllerInterface(ctrl)
	s.inputController = mock.NewMockInputControllerInterface(ctrl)
	s.history = &history.History{Data: []string{}}
	s.statementController = mock.NewMockStatementControllerInterface(ctrl)
	s.interactiveOutputController = mock.NewMockOutputControllerInterface(ctrl)
	s.basicOutputController = mock.NewMockOutputControllerInterface(ctrl)
	s.store = mock.NewMockStoreInterface(ctrl)
	s.resultFetcher = mock.NewMockResultFetcherInterface(ctrl)

	s.app = Application{
		history:                     s.history,
		store:                       s.store,
		resultFetcher:               s.resultFetcher,
		appController:               s.appController,
		inputController:             s.inputController,
		statementController:         s.statementController,
		interactiveOutputController: s.interactiveOutputController,
		basicOutputController:       s.basicOutputController,
		tokenRefreshFunc:            authenticated,
	}
}

func authenticated() error {
	return nil
}

func unauthenticated() error {
	return errors.New("401 unauthorized")
}

func manualStop() error {
	return errors.New("manual stop")
}

func (s *ApplicationTestSuite) runMainLoop(stopAfterLoopFinishes bool) string {
	if stopAfterLoopFinishes {
		// this makes the loop stop after one iteration
		s.inputController.EXPECT().GetUserInput().Return("")
		s.inputController.EXPECT().IsSpecialInput("").DoAndReturn(func(string) bool {
			s.app.tokenRefreshFunc = manualStop
			return true
		})
		s.appController.EXPECT().ExitApplication()
	}

	output := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrintLoop)
	return output
}

func (s *ApplicationTestSuite) TestReplDoesNotRunWhenUnauthenticated() {
	s.app.tokenRefreshFunc = unauthenticated
	s.appController.EXPECT().ExitApplication()

	actual := s.runMainLoop(false)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplContinuesOnSpecialInput() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(true)

	actual := s.runMainLoop(true)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplAppendsStatementToHistoryAndStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := s.runMainLoop(true)

	cupaloy.SnapshotT(s.T(), actual)
	require.Equal(s.T(), []string{userInput}, s.history.Data)
}

func (s *ApplicationTestSuite) TestReplStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := s.runMainLoop(true)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplReturnsWhenHandleStatementResultsReturnsTrue() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.resultFetcher.EXPECT().Init(statement)
	s.store.EXPECT().FetchStatementResults(statement).Return(&statement, nil)
	s.basicOutputController.EXPECT().VisualizeResults()

	actual := s.runMainLoop(true)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplDoesNotReturnWhenHandleStatementResultsReturnsFalse() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.resultFetcher.EXPECT().Init(statement)
	s.store.EXPECT().FetchStatementResults(statement).Return(&statement, nil)
	s.basicOutputController.EXPECT().VisualizeResults()

	actual := s.runMainLoop(true)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestShouldUseTView() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      types.OutputControllerInterface
	}{
		{
			name:      "local statement should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true},
			want:      s.basicOutputController,
		},
		{
			name:      "local statement should not use TView even if unbounded",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: true},
			want:      s.basicOutputController,
		},
		{
			name:      "non-local unbounded statement should always use TView",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      s.interactiveOutputController,
		},
		{
			name:      "statement with no results should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false},
			want:      s.basicOutputController,
		},
		{
			name:      "statement with empty results should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      s.basicOutputController,
		},
		{
			name: "statement with one column and two rows should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1"},
				Rows: []types.StatementResultRow{
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
				},
			}},
			want: s.basicOutputController,
		},
		{
			name: "statement with two columns and one row should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: s.basicOutputController,
		},
		{
			name: "statement with two columns and two rows should use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows: []types.StatementResultRow{
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
				},
			}},
			want: s.interactiveOutputController,
		},
		{
			name: "local statement with two columns and two rows should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows: []types.StatementResultRow{
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
				},
			}},
			want: s.basicOutputController,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, s.app.getOutputController(tt.statement))
		})
	}
}

func (s *ApplicationTestSuite) TestSyncAccessToTokenRefreshFunction() {
	dummyVariableToManipulate := 0
	numGoroutinesToSpawn := 1000
	s.app.tokenRefreshFunc = synchronizedTokenRefresh(func() error {
		dummyVariableToManipulate++
		return nil
	})
	s.testConcurrentAccess(numGoroutinesToSpawn, func() {
		_ = s.app.tokenRefreshFunc()
	})
	require.Equal(s.T(), numGoroutinesToSpawn, dummyVariableToManipulate)
}

func (s *ApplicationTestSuite) testConcurrentAccess(numGoroutinesToSpawn int, funcToExecute func()) {
	var wg sync.WaitGroup
	wg.Add(numGoroutinesToSpawn)
	for i := 0; i < numGoroutinesToSpawn; i++ {
		go func() {
			funcToExecute()
			wg.Done()
		}()
	}
	wg.Wait()
}
