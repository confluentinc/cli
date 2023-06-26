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
	inputController InputController
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
	s.inputController = InputController{
		appController:  s.appController,
		History:        s.history,
		prompt:         s.prompt,
		reverseISearch: s.reverseISearch,
	}
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

func (s *InputControllerTestSuite) TestIsSpecialInputReturnsTrueWhenReverseSearchIsEnabledAndSetsInitialBuffer() {
	s.inputController.reverseISearchEnabled = true
	searchResult := "search result"
	s.reverseISearch.EXPECT().ReverseISearch(s.history.Data).Return(searchResult)

	isSpecialInput := s.inputController.IsSpecialInput("input")

	require.False(s.T(), s.inputController.reverseISearchEnabled)
	require.Equal(s.T(), searchResult, s.inputController.InitialBuffer)
	require.True(s.T(), isSpecialInput)
}

func (s *InputControllerTestSuite) TestIsSpecialInputReturnsTrueWhenShouldExitIsTrue() {
	s.inputController.shouldExit = true
	s.appController.EXPECT().ExitApplication()

	isSpecialInput := s.inputController.IsSpecialInput("input")

	require.True(s.T(), isSpecialInput)
}

func (s *InputControllerTestSuite) TestIsSpecialInputReturnsTrueWhenUserInputIsEmpty() {
	s.appController.EXPECT().ExitApplication()

	isSpecialInput := s.inputController.IsSpecialInput("")

	require.True(s.T(), isSpecialInput)
}

func (s *InputControllerTestSuite) TestIsSpecialInputReturnsFalse() {
	isSpecialInput := s.inputController.IsSpecialInput("select 1;")

	require.False(s.T(), isSpecialInput)
}
