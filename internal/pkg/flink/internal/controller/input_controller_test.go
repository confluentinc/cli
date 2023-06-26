package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type InputControllerTestSuite struct {
	suite.Suite
	mockAppController   *mock.MockApplicationControllerInterface
	mockTableController *mock.MockTableControllerInterface
	mockPrompt          *mock.MockIPrompt
	mockStore           *mock.MockStoreInterface
	mockConsoleParser   *mock.MockConsoleParser
	mockReverseISearch  *mock.MockReverseISearch
}

func TestInputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(InputControllerTestSuite))
}

func (s *InputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockAppController = mock.NewMockApplicationControllerInterface(ctrl)
	s.mockTableController = mock.NewMockTableControllerInterface(ctrl)
	s.mockPrompt = mock.NewMockIPrompt(ctrl)
	s.mockStore = mock.NewMockStoreInterface(ctrl)
	s.mockConsoleParser = mock.NewMockConsoleParser(ctrl)
	s.mockReverseISearch = mock.NewMockReverseISearch(ctrl)
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

func (s *InputControllerTestSuite) runMainLoop(inputController *InputController, stopAfterLoopFinishes bool) string {
	if stopAfterLoopFinishes {
		// this makes the loop stop after one iteration
		s.mockPrompt.EXPECT().Input().Return("")
		s.mockAppController.EXPECT().ExitApplication().Do(func() { inputController.shouldExit = true })
	}

	output := s.runAndCaptureSTDOUT(inputController.RunInteractiveInput)
	return output
}

func (s *InputControllerTestSuite) getStdoutTable(statementResults types.StatementResults) string {
	return s.runAndCaptureSTDOUT(func() {
		rawTable := tablewriter.NewWriter(os.Stdout)
		rawTable.SetAutoFormatHeaders(false)
		rawTable.SetHeader(statementResults.Headers)
		for _, statementResultRow := range statementResults.GetRows() {
			row := make([]string, len(statementResultRow.Fields))
			for idx, field := range statementResultRow.Fields {
				row[idx] = field.ToString()
			}
			rawTable.Append(row)
		}
		rawTable.Render()
	})
}

func (s *InputControllerTestSuite) TestRenderError() {
	inputController := InputController{appController: s.mockAppController}
	err := types.StatementError{HttpResponseCode: http.StatusUnauthorized}

	// Test unauthorized error
	result := inputController.isSessionValid(err)
	require.False(s.T(), result)

	// Test other error
	err = types.StatementError{Message: "something went wrong."}
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
			statement: types.ProcessedStatement{IsLocalStatement: false},
			want:      false,
		},
		{
			name:      "statement with empty results should not use TView",
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
		statement types.ProcessedStatement
		want      string
	}{
		{
			name:      "statement with name",
			statement: types.ProcessedStatement{StatementName: "test-statement", Status: types.FAILED},
			want:      "Statement name: test-statement\nError: statement submission failed\n",
		},
		{
			name:      "statement with name and status detail",
			statement: types.ProcessedStatement{StatementName: "test-statement", Status: types.FAILED, StatusDetail: "status-detail"},
			want:      "Statement name: test-statement\nError: statement submission failed\nstatus-detail.\n",
		},
		{
			name:      "statement without name",
			statement: types.ProcessedStatement{Status: types.FAILED},
			want:      "Error: statement submission failed\n",
		},
		{
			name:      "statement without name but with status detail",
			statement: types.ProcessedStatement{Status: types.FAILED, StatusDetail: "status-detail"},
			want:      "Error: statement submission failed\nstatus-detail.\n",
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
		statement types.ProcessedStatement
		want      string
	}{
		{
			name:      "statement with name",
			statement: types.ProcessedStatement{StatementName: "test-statement", Status: types.RUNNING},
			want:      "Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n",
		},
		{
			name:      "statement with name and status detail",
			statement: types.ProcessedStatement{StatementName: "test-statement", Status: types.RUNNING, StatusDetail: "status-detail"},
			want:      "Statement name: test-statement\nStatement successfully submitted.\nFetching results...\nstatus-detail.\n",
		},
		{
			name:      "statement without name",
			statement: types.ProcessedStatement{Status: types.RUNNING},
			want:      "Statement successfully submitted.\nFetching results...\n",
		},
		{
			name:      "statement without name but with status detail",
			statement: types.ProcessedStatement{Status: types.RUNNING, StatusDetail: "status-detail"},
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
		authenticated: func() error {
			return nil
		},
	}

	s.mockPrompt.EXPECT().Input().Return("")
	s.mockAppController.EXPECT().ExitApplication().Do(func() { inputController.shouldExit = true })

	// When
	actual := s.runMainLoop(inputController, false)

	// Then
	require.Empty(s.T(), actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputExitsWhenShouldExitTrue() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		authenticated: func() error {
			return nil
		},
		shouldExit: true,
	}

	// When
	actual := s.runMainLoop(inputController, false)

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

	s.mockAppController.EXPECT().ExitApplication()

	// When
	actual := s.runMainLoop(inputController, false)

	// Then
	expected := fmt.Sprintf("Error: %s\n", inputController.authenticated().Error())
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputWithBackwardSearch() {
	// Given
	inputController := &InputController{
		appController:         s.mockAppController,
		prompt:                s.mockPrompt,
		reverseISearchEnabled: true,
		History:               &history.History{Data: []string{"select 1;"}},
		reverseISearch:        s.mockReverseISearch,
		authenticated: func() error {
			return nil
		},
	}

	searchResult := "search result"
	buffer := prompt.NewBuffer()
	s.mockPrompt.EXPECT().Input().Return("select 1;")
	s.mockReverseISearch.EXPECT().ReverseISearch(inputController.History.Data).Return(searchResult)
	s.mockPrompt.EXPECT().Buffer().Return(buffer)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	require.Equal(s.T(), "", actual)
	require.False(s.T(), inputController.reverseISearchEnabled)
	require.Equal(s.T(), searchResult, buffer.Text())
	// buffer should be reset when the next iteration starts
	require.Equal(s.T(), "", inputController.InitialBuffer)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputPrintsErrorAndContinuesOnProcessStatementError() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statementError := &types.StatementError{Message: "error"}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(nil, statementError)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	require.Equal(s.T(), fmt.Sprintf("%s\n", statementError.Error()), actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputExitsOn401FromProcessStatement() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statementError := &types.StatementError{Message: "error", HttpResponseCode: 401}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(nil, statementError)
	s.mockAppController.EXPECT().ExitApplication().Do(func() { inputController.shouldExit = true })

	// When
	actual := s.runMainLoop(inputController, false)

	// Then
	require.Equal(s.T(), fmt.Sprintf("%s\n", statementError.Error()), actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputStoresInputInHistory() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statementError := &types.StatementError{Message: "error"}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(nil, statementError)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	require.Equal(s.T(), fmt.Sprintf("%s\n", statementError.Error()), actual)
	require.Equal(s.T(), inputController.History.Data, []string{input})
}

func (s *InputControllerTestSuite) TestRunInteractiveInputPrintsErrorAndContinuesOnWaitPendingStatementError() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	statementError := &types.StatementError{Message: "error", FailureMessage: "error details"}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).Return(nil, statementError)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s\n", statementError.Error())
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputCancelsAndDeletesStatementOnUserInterrupt() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	statementError := &types.StatementError{Message: "result retrieval aborted. Statement will be deleted", HttpResponseCode: 499}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return([]byte{byte(prompt.ControlC)}, nil)
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).DoAndReturn(
		func(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
			time.Sleep(1 * time.Second)
			return nil, statementError
		})
	s.mockStore.EXPECT().DeleteStatement(statement.StatementName)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s\n", statementError.Error())
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputExitsOn401FromWaitPendingStatement() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	statementError := &types.StatementError{Message: "error", HttpResponseCode: 401}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).Return(nil, statementError)
	s.mockAppController.EXPECT().ExitApplication().Do(func() { inputController.shouldExit = true })

	// When
	actual := s.runMainLoop(inputController, false)

	// Then
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s\n", statementError.Error())
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputPrintsErrorAndContinuesOnFetchStatementResultsError() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	completedStatement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.COMPLETED,
		StatusDetail:  "status-detail",
	}
	statementError := &types.StatementError{Message: "error"}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).Return(&completedStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(completedStatement).Return(nil, statementError)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s.\n%s\n", completedStatement.StatusDetail, statementError.Error())
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputShouldOpenTView() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		table:         s.mockTableController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	completedStatement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.COMPLETED,
		StatusDetail:  "status-detail",
		PageToken:     "not-empty",
	}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).Return(&completedStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)
	s.mockTableController.EXPECT().Init(completedStatement)

	// When
	actual := s.runMainLoop(inputController, false)

	// Then
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s.\n", completedStatement.StatusDetail)
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputShouldPrintNoRows() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		table:         s.mockTableController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	completedStatement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.COMPLETED,
		StatusDetail:  "status-detail",
	}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).Return(&completedStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s.\n\nThe server returned empty rows for this statement.\n", completedStatement.StatusDetail)
	require.Equal(s.T(), expected, actual)
}

func (s *InputControllerTestSuite) TestRunInteractiveInputShouldPrintTable() {
	// Given
	inputController := &InputController{
		appController: s.mockAppController,
		table:         s.mockTableController,
		prompt:        s.mockPrompt,
		store:         s.mockStore,
		consoleParser: s.mockConsoleParser,
		History:       &history.History{},
		authenticated: func() error {
			return nil
		},
	}

	input := "select 1;"
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	completedStatement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.COMPLETED,
		StatusDetail:  "status-detail",
		StatementResults: &types.StatementResults{
			Headers: []string{"column"},
			Rows: []types.StatementResultRow{
				{
					Operation: 0,
					Fields: []types.StatementResultField{
						types.AtomicStatementResultField{
							Type:  "INTEGER",
							Value: "0",
						},
					},
				},
			},
		},
	}
	s.mockPrompt.EXPECT().Input().Return(input)
	s.mockStore.EXPECT().ProcessStatement(input).Return(&statement, nil)
	s.mockConsoleParser.EXPECT().Read().Return(nil, nil).AnyTimes()
	s.mockStore.EXPECT().WaitPendingStatement(gomock.Any(), statement).Return(&completedStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(completedStatement).Return(&completedStatement, nil)

	// When
	actual := s.runMainLoop(inputController, true)

	// Then
	table := s.getStdoutTable(*completedStatement.StatementResults)
	expected := fmt.Sprintf("Statement name: test-statement\nStatement successfully submitted.\nFetching results...\n%s.\n%s", completedStatement.StatusDetail, table)
	require.Equal(s.T(), expected, actual)
}
