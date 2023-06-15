package controller

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type InputControllerTestSuite struct {
	suite.Suite
	mockAppController   *mock.MockApplicationControllerInterface
	mockTableController *mock.MockTableControllerInterface
	mockPrompt          *mock.MockIPrompt
	mockStore           *mock.MockStoreInterface
}

func TestInputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(InputControllerTestSuite))
}

func (s *InputControllerTestSuite) runAndCaptureSTDOUT(test func()) string {
	// Redirect STDOUT to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the test
	test()

	// Close the writer and restore the original STDOUT
	err := w.Close()
	require.NoError(s.T(), err)
	os.Stdout = old

	// Read the output from the buffer
	output := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output <- string(buf[:n])
	}()
	return <-output
}

func (s *InputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockAppController = mock.NewMockApplicationControllerInterface(ctrl)
	s.mockTableController = mock.NewMockTableControllerInterface(ctrl)
	s.mockPrompt = mock.NewMockIPrompt(ctrl)
	s.mockStore = mock.NewMockStoreInterface(ctrl)
}

func (s *InputControllerTestSuite) TestRenderError() {
	inputController := &InputController{appController: s.mockAppController}
	err := &types.StatementError{HttpResponseCode: http.StatusUnauthorized}

	// Test unauthorized error - should exit application
	s.mockAppController.EXPECT().ExitApplication()
	result := inputController.isSessionValid(err)
	require.False(s.T(), result)

	// Test other error
	err = &types.StatementError{Message: "something went wrong."}
	result = inputController.isSessionValid(err)
	require.True(s.T(), result)
}

func (s *InputControllerTestSuite) TestShouldUseTView() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      bool
	}{
		{
			name:      "local statement should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true},
			want:      false,
		},
		{
			name:      "local statement should not use TView even if unbounded",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: true},
			want:      false,
		},
		{
			name:      "non-local unbounded statement should always use TView",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      true,
		},
		{
			name:      "statement with no results should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      false,
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
			want: false,
		},
		{
			name: "statement with two columns and one row should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: false,
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
			want: true,
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
			want: false,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldUseTView(tt.statement))
		})
	}
}

func (s *InputControllerTestSuite) TestRenderMsgAndStatusLocalStatements() {
	tests := []struct {
		name      string
		statement *types.ProcessedStatement
		want      string
	}{
		{
			name:      "nil",
			statement: nil,
			want:      "",
		},
		{
			name:      "local failed statement",
			statement: &types.ProcessedStatement{IsLocalStatement: true, Status: types.FAILED},
			want:      "Error: Couldn't process statement. Please check your statement and try again\n",
		},
		{
			name:      "local non-failed statement",
			statement: &types.ProcessedStatement{IsLocalStatement: true, Status: types.RUNNING},
			want:      "Statement successfully submitted.\n",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			actual := s.runAndCaptureSTDOUT(func() {
				renderMsgAndStatus(tt.statement)
			})
			require.Equal(t, tt.want, actual)
		})
	}
}

func (s *InputControllerTestSuite) TestRenderMsgAndStatusNonLocalFailedStatements() {
	tests := []struct {
		name      string
		statement *types.ProcessedStatement
		want      string
	}{
		{
			name:      "nil",
			statement: nil,
			want:      "",
		},
		{
			name:      "statement with name",
			statement: &types.ProcessedStatement{StatementName: "test-statement", Status: types.FAILED},
			want:      "Statement name: test-statement\nError: Statement submission failed. There could a problem with the server right now. Check your statement and try again\n",
		},
		{
			name:      "statement with name and status detail",
			statement: &types.ProcessedStatement{StatementName: "test-statement", Status: types.FAILED, StatusDetail: "status-detail"},
			want:      "Statement name: test-statement\nError: Statement submission failed. There could a problem with the server right now. Check your statement and try again\nstatus-detail.\n",
		},
		{
			name:      "statement without name",
			statement: &types.ProcessedStatement{Status: types.FAILED},
			want:      "Error: Statement submission failed. There could a problem with the server right now. Check your statement and try again\n",
		},
		{
			name:      "statement without name but with status detail",
			statement: &types.ProcessedStatement{Status: types.FAILED, StatusDetail: "status-detail"},
			want:      "Error: Statement submission failed. There could a problem with the server right now. Check your statement and try again\nstatus-detail.\n",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			actual := s.runAndCaptureSTDOUT(func() {
				renderMsgAndStatus(tt.statement)
			})
			require.Equal(t, tt.want, actual)
		})
	}
}

func (s *InputControllerTestSuite) TestRenderMsgAndStatusNonLocalNonFailedStatements() {
	tests := []struct {
		name      string
		statement *types.ProcessedStatement
		want      string
	}{
		{
			name:      "nil",
			statement: nil,
			want:      "",
		},
		{
			name:      "statement with name",
			statement: &types.ProcessedStatement{StatementName: "test-statement", Status: types.RUNNING},
			want:      "Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n",
		},
		{
			name:      "statement with name and status detail",
			statement: &types.ProcessedStatement{StatementName: "test-statement", Status: types.RUNNING, StatusDetail: "status-detail"},
			want:      "Statement name: test-statement\nStatement successfully submitted.\nFetching results...\nstatus-detail.\n",
		},
		{
			name:      "statement without name",
			statement: &types.ProcessedStatement{Status: types.RUNNING},
			want:      "Statement successfully submitted.\nFetching results...\n",
		},
		{
			name:      "statement without name but with status detail",
			statement: &types.ProcessedStatement{Status: types.RUNNING, StatusDetail: "status-detail"},
			want:      "Statement successfully submitted.\nFetching results...\nstatus-detail.\n",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			actual := s.runAndCaptureSTDOUT(func() {
				renderMsgAndStatus(tt.statement)
			})
			require.Equal(t, tt.want, actual)
		})
	}
}

func (s *InputControllerTestSuite) TestRunInteractiveInputExitsWhenEmptyPromptReturn() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		shouldExit:    false,
	}

	s.mockPrompt.EXPECT().Input().Return("")
	s.mockAppController.EXPECT().ExitApplication()

	// When
	actual := s.runAndCaptureSTDOUT(inputController.RunInteractiveInput)

	// Then
	require.Empty(s.T(), actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputExitsWhenShouldExitTrue() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		shouldExit:    true,
	}

	s.mockPrompt.EXPECT().Input().Return("select 1;")
	s.mockAppController.EXPECT().ExitApplication()

	// When
	actual := s.runAndCaptureSTDOUT(inputController.RunInteractiveInput)

	// Then
	require.Empty(s.T(), actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputExitsWhenNotAuthenticated() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		authenticated: func() error {
			return errors.New("401 unauthorized")
		},
	}

	s.mockPrompt.EXPECT().Input().Return("select 1;")
	s.mockAppController.EXPECT().ExitApplication()

	// When
	actual := s.runAndCaptureSTDOUT(inputController.RunInteractiveInput)

	// Then
	expected := fmt.Sprintf("%s\n", inputController.authenticated().Error())
	require.Equal(s.T(), expected, actual)
}
