package controller

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableControllerTestSuite struct {
	suite.Suite
	tableController *TableController
	fetchController *mock.MockFetchControllerInterface
	dummyTViewApp   *tview.Application
}

func TestTableControllerTestSuite(t *testing.T) {
	suite.Run(t, new(TableControllerTestSuite))
}

func (s *TableControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.fetchController = mock.NewMockFetchControllerInterface(ctrl)
	s.dummyTViewApp = tview.NewApplication()
	s.tableController = NewTableController(s.fetchController).(*TableController)
}

func (s *TableControllerTestSuite) TestCloseTableViewOnUserInput() {
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Q", input: tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)},
		{name: "Test CtrlQ", input: tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)},
		{name: "Test Escape", input: tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			// Expected mock calls
			s.fetchController.EXPECT().Close()

			// When
			result := s.tableController.inputCapture(testCase.input)

			// Then
			require.Nil(s.T(), result)
		})
	}
}

func (s *TableControllerTestSuite) TestToggleTableModeOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.tableController.Init(types.ProcessedStatement{})
	input := tcell.NewEventKey(tcell.KeyRune, 'M', tcell.ModNone)
	s.fetchController.EXPECT().ToggleTableMode()
	s.updateTableMockCalls()

	result := s.tableController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) updateTableMockCalls() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Paused)
	s.fetchController.EXPECT().GetResults().Return(&types.MaterializedStatementResults{})
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
}

func (s *TableControllerTestSuite) TestToggleRefreshResultsOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.tableController.Init(types.ProcessedStatement{})
	input := tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone)
	s.fetchController.EXPECT().ToggleAutoRefresh()
	s.updateTableMockCalls()

	result := s.tableController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestFetchNextPageOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.tableController.Init(types.ProcessedStatement{})
	input := tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone)
	s.fetchController.EXPECT().FetchNextPage()
	s.updateTableMockCalls()

	result := s.tableController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestJumpToLastPageOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.tableController.Init(types.ProcessedStatement{})
	input := tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
	s.fetchController.EXPECT().JumpToLastPage()
	s.updateTableMockCalls()

	result := s.tableController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestNonSupportedUserInput() {
	// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
	// When we return the event, it's forwarded to tview
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)

	result := s.tableController.inputCapture(input)

	require.Equal(s.T(), input, result)
}

func (s *TableControllerTestSuite) TestOpenRowViewOnUserInput() {
	// Given
	materializedStatementResults := getResultsExample()
	s.initMockCalls(materializedStatementResults)
	s.tableController.Init(types.ProcessedStatement{})

	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
	s.fetchController.EXPECT().GetResults().Return(materializedStatementResults)

	// When
	result := s.tableController.inputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

	// Then
	require.Nil(s.T(), result)
	require.True(s.T(), s.tableController.isRowViewOpen)
	// last row should be selected
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableController.tableView.GetSelectedRow())
}

func getResultsExample() *types.MaterializedStatementResults {
	materializedStatementResults := types.NewMaterializedStatementResults([]string{"Count"}, 10)
	for i := 0; i < 10; i++ {
		materializedStatementResults.Append(types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(i),
				},
			},
		})
	}
	return &materializedStatementResults
}

func (s *TableControllerTestSuite) initMockCalls(materializedStatementResults *types.MaterializedStatementResults) {
	s.fetchController.EXPECT().Init(gomock.Any())
	s.fetchController.EXPECT().SetAutoRefreshCallback(gomock.Any())
	s.fetchController.EXPECT().GetResults().Return(materializedStatementResults)
	s.fetchController.EXPECT().IsTableMode().Return(materializedStatementResults.IsTableMode())
	s.fetchController.EXPECT().GetFetchState().Return(types.Paused)
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
}

func (s *TableControllerTestSuite) TestCloseRowViewOnUserInput() {
	// Given
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Q", input: tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)},
		{name: "Test CtrlQ", input: tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)},
		{name: "Test Escape", input: tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			s.initMockCalls(&types.MaterializedStatementResults{})
			s.tableController.Init(types.ProcessedStatement{})
			s.tableController.isRowViewOpen = true

			// When
			result := s.tableController.inputCapture(testCase.input)

			// Then
			require.Nil(s.T(), result)
			require.False(s.T(), s.tableController.isRowViewOpen)
		})
	}
}

func (s *TableControllerTestSuite) TestNonSupportedUserInputInRowView() {
	// Given
	s.tableController.isRowViewOpen = true

	// When
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := s.tableController.inputCapture(input)

	// Then
	require.NotNil(s.T(), result)
	require.Equal(s.T(), input, result)
	require.True(s.T(), s.tableController.isRowViewOpen)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysTableMode() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Failed)

	actual := s.tableController.getTableTitle()

	require.Contains(s.T(), actual, "Table mode")
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysChangelogMode() {
	s.fetchController.EXPECT().IsTableMode().Return(false)
	s.fetchController.EXPECT().GetFetchState().Return(types.Failed)

	actual := s.tableController.getTableTitle()

	require.Contains(s.T(), actual, "Changelog mode")
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysComplete() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Completed)

	actual := s.tableController.getTableTitle()

	require.Contains(s.T(), actual, "(completed)")
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysFailed() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Failed)

	actual := s.tableController.getTableTitle()

	require.Contains(s.T(), actual, "(auto refresh failed)")
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysPaused() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Paused)

	actual := s.tableController.getTableTitle()

	require.Contains(s.T(), actual, "(auto refresh paused)")
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysRunning() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Running)

	actual := s.tableController.getTableTitle()

	require.Contains(s.T(), actual, fmt.Sprintf("(auto refresh %vs)", results.DefaultRefreshInterval/1000))
}
