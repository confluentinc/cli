package controller

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
)

type InputControllerTestSuite struct {
	suite.Suite
	inputController *InputController
	appController   *mock.MockApplicationControllerInterface
	history         *history.History
	prompt          *mock.MockIPrompt
	reverseISearch  *mock.MockReverseISearch
}

func TestInputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(InputControllerTestSuite))
}

func (s *InputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.appController = mock.NewMockApplicationControllerInterface(ctrl)
	s.history = &history.History{Data: []string{}}
	s.prompt = mock.NewMockIPrompt(ctrl)
	s.reverseISearch = mock.NewMockReverseISearch(ctrl)
	s.inputController = NewInputController(s.history, nil).(*InputController)
	s.inputController.reverseISearch = s.reverseISearch
	s.inputController.prompt = s.prompt
}

func (s *InputControllerTestSuite) TestGetUserInput() {
	input := "input"
	s.prompt.EXPECT().Input().Return(input)

	actual := s.inputController.GetUserInput()

	require.Equal(s.T(), input, actual)
}

func (s *InputControllerTestSuite) TestGetUserInputSetsInitialBuffer() {
	s.inputController.InitialBuffer = "not-empty"
	input := fmt.Sprintf("%s %s", s.inputController.InitialBuffer, "input")
	buffer := prompt.NewBuffer()
	s.prompt.EXPECT().Buffer().Return(buffer)
	s.prompt.EXPECT().Input().Return(input)

	actual := s.inputController.GetUserInput()

	require.Equal(s.T(), buffer.Text(), "not-empty")
	require.Equal(s.T(), input, actual)
}

func (s *InputControllerTestSuite) TestHasUserEnabledReverseSearchShouldBeTrue() {
	s.inputController.reverseISearchEnabled = true

	hasUserEnabledReverseSearch := s.inputController.HasUserEnabledReverseSearch()

	require.True(s.T(), hasUserEnabledReverseSearch)
}

func (s *InputControllerTestSuite) TestHasUserEnabledReverseSearchShouldBeFalse() {
	s.inputController.reverseISearchEnabled = false

	hasUserEnabledReverseSearch := s.inputController.HasUserEnabledReverseSearch()

	require.False(s.T(), hasUserEnabledReverseSearch)
}

func (s *InputControllerTestSuite) TestStartReverseSearch() {
	searchResult := "search result"
	s.reverseISearch.EXPECT().ReverseISearch(s.history.Data).Return(searchResult)

	s.inputController.StartReverseSearch()

	require.False(s.T(), s.inputController.reverseISearchEnabled)
	require.Equal(s.T(), searchResult, s.inputController.InitialBuffer)
}

func (s *InputControllerTestSuite) TestHasUserInitiatedExitShouldBeTrueWhenShouldExitIsTrue() {
	s.inputController.shouldExit = true

	hasUserInitiatedExit := s.inputController.HasUserInitiatedExit("exit;")

	require.True(s.T(), hasUserInitiatedExit)
}

func (s *InputControllerTestSuite) TestHasUserInitiatedExitShouldBeTrueWhenUserInputEmpty() {
	hasUserInitiatedExit := s.inputController.HasUserInitiatedExit("")

	require.True(s.T(), hasUserInitiatedExit)
}

func (s *InputControllerTestSuite) TestHasUserInitiatedExitShouldBeFalse() {
	s.inputController.reverseISearchEnabled = false

	hasUserEnabledReverseSearch := s.inputController.HasUserEnabledReverseSearch()

	require.False(s.T(), hasUserEnabledReverseSearch)
}

func (s *InputControllerTestSuite) TestTurnOnSmartCompletion() {
	s.inputController.smartCompletion = false

	s.inputController.toggleSmartCompletion()

	require.True(s.T(), s.inputController.smartCompletion)
}

func (s *InputControllerTestSuite) TestTurnOffSmartCompletion() {
	s.inputController.smartCompletion = true

	s.inputController.toggleSmartCompletion()

	require.False(s.T(), s.inputController.smartCompletion)
}
