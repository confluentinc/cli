package app

import (
	"sync"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/history"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v3/pkg/flink/test"
	"github.com/confluentinc/cli/v3/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
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
		refreshToken:                authenticated,
	}
}

func authenticated() error {
	return nil
}

func unauthenticated() error {
	return errors.New("401 unauthorized")
}

func (s *ApplicationTestSuite) TestReplDoesNotRunWhenUnauthenticated() {
	s.app.refreshToken = unauthenticated
	s.appController.EXPECT().ExitApplication()

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrintLoop)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplContinuesWhenUserEnabledReverseSearch() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().HasUserEnabledReverseSearch().Return(true)
	s.inputController.EXPECT().StartReverseSearch()

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrint)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplExitsAppWhenUserInitiatedExit() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().HasUserEnabledReverseSearch().Return(false)
	s.inputController.EXPECT().HasUserInitiatedExit(userInput).Return(true)
	s.appController.EXPECT().ExitApplication()

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrint)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplAppendsStatementToHistoryAndStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().HasUserEnabledReverseSearch().Return(false)
	s.inputController.EXPECT().HasUserInitiatedExit(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrint)

	cupaloy.SnapshotT(s.T(), actual)
	require.Equal(s.T(), []string{userInput}, s.history.Data)
}

func (s *ApplicationTestSuite) TestReplStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().HasUserEnabledReverseSearch().Return(false)
	s.inputController.EXPECT().HasUserInitiatedExit(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrint)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplUsesBasicOutput() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().HasUserEnabledReverseSearch().Return(false)
	s.inputController.EXPECT().HasUserInitiatedExit(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.resultFetcher.EXPECT().Init(statement)
	s.basicOutputController.EXPECT().VisualizeResults()

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrint)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestReplUsesInteractiveOutput() {
	userInput := "test-input"
	statement := types.ProcessedStatement{PageToken: "not-empty"}
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().HasUserEnabledReverseSearch().Return(false)
	s.inputController.EXPECT().HasUserInitiatedExit(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.resultFetcher.EXPECT().Init(statement)
	s.interactiveOutputController.EXPECT().VisualizeResults()

	actual := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrint)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *ApplicationTestSuite) TestShouldUseTView() {
	app := Application{
		interactiveOutputController: &controller.InteractiveOutputController{},
		basicOutputController:       &controller.BasicOutputController{},
	}
	tests := []struct {
		name          string
		statement     types.ProcessedStatement
		isBasicOutput bool
	}{
		{
			name:          "local statement should not use TView",
			statement:     types.ProcessedStatement{IsLocalStatement: true},
			isBasicOutput: true,
		},
		{
			name:          "local statement should not use TView even if unbounded",
			statement:     types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: true},
			isBasicOutput: true,
		},
		{
			name: "local statement should not use TView even if unbounded and more than 3 columns",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: true, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2", "Column 3", "Column 4"},
				Rows:    []types.StatementResultRow{},
			}},
			isBasicOutput: true,
		},
		{
			name:          "non-local unbounded statement should always use TView",
			statement:     types.ProcessedStatement{PageToken: "NOT_EMPTY", StatementResults: &types.StatementResults{}},
			isBasicOutput: false,
		},
		{
			name:          "select statement should always use TView",
			statement:     types.ProcessedStatement{IsSelectStatement: true, StatementResults: &types.StatementResults{}},
			isBasicOutput: false,
		},
		{
			name:          "statement with no results should not use TView",
			statement:     types.ProcessedStatement{},
			isBasicOutput: true,
		},
		{
			name:          "statement with empty results should not use TView",
			statement:     types.ProcessedStatement{StatementResults: &types.StatementResults{}},
			isBasicOutput: true,
		},
		{
			name: "statement with 3 columns should not use TView",
			statement: types.ProcessedStatement{StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2", "Column 3"},
				Rows:    []types.StatementResultRow{},
			}},
			isBasicOutput: true,
		},
		{
			name: "statement with 4 columns should not use TView",
			statement: types.ProcessedStatement{StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2", "Column 3", "Column 4"},
				Rows:    []types.StatementResultRow{},
			}},
			isBasicOutput: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			if tt.isBasicOutput {
				actual, ok := app.getOutputController(tt.statement).(*controller.BasicOutputController)
				require.True(t, ok)
				require.Equal(t, app.basicOutputController, actual)
				return
			}

			actual, ok := app.getOutputController(tt.statement).(*controller.InteractiveOutputController)
			require.True(t, ok)
			require.Equal(t, app.interactiveOutputController, actual)
		})
	}
}

func (s *ApplicationTestSuite) TestSyncAccessToTokenRefreshFunction() {
	dummyVariableToManipulate := 0
	numGoroutinesToSpawn := 1000
	s.app.refreshToken = synchronizedTokenRefresh(func() error {
		dummyVariableToManipulate++
		return nil
	})
	s.testConcurrentAccess(numGoroutinesToSpawn, func() {
		_ = s.app.refreshToken()
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

func (s *ApplicationTestSuite) TestPanicRecovery() {
	// Given
	callCount := 0
	s.app.reportUsage = func() {
		callCount++
	}
	s.inputController.EXPECT().GetUserInput().Do(func() {
		panic("err in repl")
	})
	s.statementController.EXPECT().CleanupStatement()

	// When
	actual := test.RunAndCaptureSTDOUT(s.T(), utils.WithCustomPanicRecovery(s.app.readEvalPrint, s.app.panicRecovery))

	// Then
	cupaloy.SnapshotT(s.T(), actual)
	require.Equal(s.T(), 1, callCount)
}
