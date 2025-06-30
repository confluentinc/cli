package controller

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v4/pkg/flink/internal/history"
	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
)

type InputControllerTestSuite struct {
	suite.Suite
	inputController *InputController
	appController   *mock.MockApplicationControllerInterface
	history         *history.History
	prompt          *mock.MockIPrompt
	reverseISearch  *mock.MockReverseISearch
	handlerCh       chan *jsonrpc2.Request
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
	s.handlerCh = make(chan *jsonrpc2.Request)
	s.inputController = NewInputController(s.history, nil, s.handlerCh, true).(*InputController)
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
	s.prompt.EXPECT().Buffer().Return(buffer).Times(5)
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
	s.reverseISearch.EXPECT().ReverseISearch(s.history.Data, "").Return(searchResult)
	s.prompt.EXPECT().Buffer().Return(prompt.NewBuffer())

	s.inputController.StartReverseSearch()

	require.False(s.T(), s.inputController.reverseISearchEnabled)
	require.Equal(s.T(), searchResult, s.inputController.InitialBuffer)
}

func (s *InputControllerTestSuite) TestSetDiagnostics() {
	diagnostics := []lsp.Diagnostic{{
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 10},
			End:   lsp.Position{Line: 0, Character: 13},
		},
		Severity: 1,
		Code:     "1234",
		Source:   "mock source",
		Message:  "Error: this is a lsp diagnostic",
	}}
	publishDiagnosticsParams := lsp.PublishDiagnosticsParams{
		URI:         "file:///tmp/test.sql",
		Diagnostics: diagnostics,
	}

	diagnosticsParams, _ := json.Marshal(publishDiagnosticsParams)
	rawParams := json.RawMessage(diagnosticsParams)
	req := &jsonrpc2.Request{
		Method: "textDocument/publishDiagnostics",
		Params: &rawParams,
	}

	s.prompt.EXPECT().SetDiagnostics(diagnostics)
	s.handlerCh <- req
	time.Sleep(100 * time.Millisecond)
}

func (s *InputControllerTestSuite) TestHasUserInitiatedExitShouldBeTrueWhenShouldExitIsTrue() {
	s.inputController.shouldExit = true

	hasUserInitiatedExit := s.inputController.HasUserInitiatedExit("exit;")

	require.True(s.T(), hasUserInitiatedExit)
}

func (s *InputControllerTestSuite) TestHasUserInitiatedQuitShouldBeTrueWhenShouldExitIsTrue() {
	s.inputController.shouldExit = true

	hasUserInitiatedExit := s.inputController.HasUserInitiatedExit("quit;")

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

func (s *InputControllerTestSuite) TestTurnOnCompletions() {
	s.inputController.completionsEnabled = false

	s.inputController.toggleCompletions()

	require.True(s.T(), s.inputController.completionsEnabled)
}

func (s *InputControllerTestSuite) TestTurnOffCompletions() {
	s.inputController.completionsEnabled = true

	s.inputController.toggleCompletions()

	require.False(s.T(), s.inputController.completionsEnabled)
}

func (s *InputControllerTestSuite) TestTurnOnDiagnostics() {
	s.inputController.diagnosticsEnabled = false
	s.prompt.EXPECT().SetDiagnostics(nil)

	s.inputController.toggleDiagnostics()

	require.True(s.T(), s.inputController.diagnosticsEnabled)
}

func (s *InputControllerTestSuite) TestTurnOffDiagnostics() {
	s.inputController.diagnosticsEnabled = true
	s.prompt.EXPECT().SetDiagnostics(nil)

	s.inputController.toggleDiagnostics()

	require.False(s.T(), s.inputController.diagnosticsEnabled)
}
