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
	app                     Application
	mockAppController       *mock.MockApplicationControllerInterface
	mockInputController     *mock.MockInputControllerInterface
	mockStatementController *mock.MockStatementControllerInterface
	mockResultsController   *mock.MockOutputControllerInterface
	history                 *history.History
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockAppController = mock.NewMockApplicationControllerInterface(ctrl)
	s.mockInputController = mock.NewMockInputControllerInterface(ctrl)
	s.history = &history.History{Data: []string{}}
	s.mockStatementController = mock.NewMockStatementControllerInterface(ctrl)
	s.mockResultsController = mock.NewMockOutputControllerInterface(ctrl)

	s.app = Application{
		history:             s.history,
		appController:       s.mockAppController,
		inputController:     s.mockInputController,
		statementController: s.mockStatementController,
		resultsController:   s.mockResultsController,
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
		s.mockInputController.EXPECT().GetUserInput().Return("")
		s.mockInputController.EXPECT().IsSpecialInput("").DoAndReturn(func(string) bool {
			s.app.authenticated = manualStop
			return true
		})
		s.mockAppController.EXPECT().ExitApplication()
	}

	output := test.RunAndCaptureSTDOUT(s.T(), s.app.readEvalPrintLoop)
	return output
}

func (s *ApplicationTestSuite) TestReplDoesNotRunWhenUnauthenticated() {
	s.app.authenticated = unauthenticated
	s.mockAppController.EXPECT().ExitApplication()

	actual := s.runMainLoop(false)

	require.Equal(s.T(), fmt.Sprintf("Error: %v\n", unauthenticated()), actual)
}

func (s *ApplicationTestSuite) TestReplContinuesOnSpecialInput() {
	userInput := "test-input"
	s.mockInputController.EXPECT().GetUserInput().Return(userInput)
	s.mockInputController.EXPECT().IsSpecialInput(userInput).Return(true)

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}

func (s *ApplicationTestSuite) TestReplAppendsStatementToHistoryAndStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.mockInputController.EXPECT().GetUserInput().Return(userInput)
	s.mockInputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.mockStatementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
	require.Equal(s.T(), []string{userInput}, s.history.Data)
}

func (s *ApplicationTestSuite) TestReplStopsOnExecuteStatementError() {
	userInput := "test-input"
	s.mockInputController.EXPECT().GetUserInput().Return(userInput)
	s.mockInputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.mockStatementController.EXPECT().ExecuteStatement(userInput).Return(nil, &types.StatementError{})

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}

func (s *ApplicationTestSuite) TestReplReturnsWhenHandleStatementResultsReturnsTrue() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	windowSize := 10
	s.mockInputController.EXPECT().GetUserInput().Return(userInput)
	s.mockInputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.mockStatementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.mockInputController.EXPECT().GetWindowWidth().Return(windowSize)
	s.mockResultsController.EXPECT().HandleStatementResults(statement, windowSize).Return(true)

	actual := s.runMainLoop(false)

	require.Equal(s.T(), "", actual)
}

func (s *ApplicationTestSuite) TestReplDoesNotReturnWhenHandleStatementResultsReturnsFalse() {
	userInput := "test-input"
	statement := types.ProcessedStatement{}
	windowSize := 10
	s.mockInputController.EXPECT().GetUserInput().Return(userInput)
	s.mockInputController.EXPECT().IsSpecialInput(userInput).Return(false)
	s.mockStatementController.EXPECT().ExecuteStatement(userInput).Return(&statement, nil)
	s.mockInputController.EXPECT().GetWindowWidth().Return(windowSize)
	s.mockResultsController.EXPECT().HandleStatementResults(statement, windowSize).Return(false)

	actual := s.runMainLoop(true)

	s.requireManuallyStopped(actual)
}
