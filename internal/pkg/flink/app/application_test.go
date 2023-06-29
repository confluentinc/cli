package app

import (
	"fmt"
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
	app                 Application
	appController       *mock.MockApplicationControllerInterface
	inputController     *mock.MockInputControllerInterface
	statementController *mock.MockStatementControllerInterface
	outputController    *mock.MockOutputControllerInterface
	history             *history.History
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
	s.outputController = mock.NewMockOutputControllerInterface(ctrl)

	s.app = Application{
		history:             s.history,
		appController:       s.appController,
		inputController:     s.inputController,
		statementController: s.statementController,
		resultsController:   s.outputController,
		authenticated:       authenticated,
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

func (s *ApplicationTestSuite) requireManuallyStopped(actual string) {
	require.Equal(s.T(), fmt.Sprintf("Error: %v\n", manualStop()), actual)
}

func (s *ApplicationTestSuite) runMainLoop(stopAfterLoopFinishes bool) string {
	if stopAfterLoopFinishes {
		// this makes the loop stop after one iteration
		s.inputController.EXPECT().GetUserInput().Return("")
		s.inputController.EXPECT().IsSpecialInput("").DoAndReturn(func(string) bool {
			s.app.authenticated = manualStop
			return true
		})
		s.appController.EXPECT().ExitApplication()
	}

	output := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrintLoop)
	return output
}

func (s *ApplicationTestSuite) TestReplDoesNotRunWhenUnauthenticated() {
	s.app.authenticated = unauthenticated
	s.appController.EXPECT().ExitApplication()

	actual := s.runMainLoop(false)

	require.Equal(s.T(), fmt.Sprintf("Error: %v\n", unauthenticated()), actual)
}

func (s *ApplicationTestSuite) TestReplContinuesOnSpecialInput() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(true)

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}

func (s *ApplicationTestSuite) TestReplAppendsStatementToHistoryAndStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
	require.Equal(s.T(), []string{userInput}, s.history.Data)
}

func (s *ApplicationTestSuite) TestReplStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}

func (s *ApplicationTestSuite) TestReplReturnsWhenHandleStatementResultsReturnsTrue() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	windowSize := 10
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.inputController.EXPECT().GetWindowWidth().Return(windowSize)
	s.outputController.EXPECT().HandleStatementResults(statement, windowSize)

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}

func (s *ApplicationTestSuite) TestReplDoesNotReturnWhenHandleStatementResultsReturnsFalse() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	windowSize := 10
	s.inputController.EXPECT().GetUserInput().Return(userInput)
	s.inputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.statementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.inputController.EXPECT().GetWindowWidth().Return(windowSize)
	s.outputController.EXPECT().HandleStatementResults(statement, windowSize)

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}
