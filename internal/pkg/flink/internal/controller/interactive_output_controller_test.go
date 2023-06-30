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

type InteractiveOutputControllerTestSuite struct {
	suite.Suite
	interactiveOutputController *InteractiveOutputController
	resultFetcher               *mock.MockResultFetcherInterface
	dummyTViewApp               *tview.Application
}

func TestInteractiveOutputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(InteractiveOutputControllerTestSuite))
}

func (s *InteractiveOutputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.resultFetcher = mock.NewMockResultFetcherInterface(ctrl)
	s.dummyTViewApp = tview.NewApplication()
	s.interactiveOutputController = NewInteractiveOutputController(s.resultFetcher).(*InteractiveOutputController)
}

func (s *InteractiveOutputControllerTestSuite) TestCloseTableViewOnUserInput() {
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
			s.resultFetcher.EXPECT().Close()

			// When
			result := s.interactiveOutputController.inputCapture(testCase.input)

			// Then
			require.Nil(s.T(), result)
		})
	}
}

func (s *InteractiveOutputControllerTestSuite) TestToggleTableModeOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.interactiveOutputController.init()
	input := tcell.NewEventKey(tcell.KeyRune, 'M', tcell.ModNone)
	s.resultFetcher.EXPECT().ToggleTableMode()
	s.updateTableMockCalls()

	result := s.interactiveOutputController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *InteractiveOutputControllerTestSuite) updateTableMockCalls() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Paused)
	s.resultFetcher.EXPECT().GetResults().Return(&types.MaterializedStatementResults{})
	s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)
}

func (s *InteractiveOutputControllerTestSuite) TestToggleRefreshResultsOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.interactiveOutputController.init()
	input := tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone)
	s.resultFetcher.EXPECT().ToggleAutoRefresh()
	s.updateTableMockCalls()

	result := s.interactiveOutputController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *InteractiveOutputControllerTestSuite) TestFetchNextPageOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.interactiveOutputController.init()
	input := tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone)
	s.resultFetcher.EXPECT().FetchNextPageAndUpdateState()
	s.updateTableMockCalls()

	result := s.interactiveOutputController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *InteractiveOutputControllerTestSuite) TestJumpToLastPageOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.interactiveOutputController.init()
	input := tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
	s.resultFetcher.EXPECT().JumpToLastPage()
	s.updateTableMockCalls()

	result := s.interactiveOutputController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *InteractiveOutputControllerTestSuite) TestNonSupportedUserInput() {
	// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
	// When we return the event, it's forwarded to tview
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)

	result := s.interactiveOutputController.inputCapture(input)

	require.Equal(s.T(), input, result)
}

func (s *InteractiveOutputControllerTestSuite) TestOpenRowViewOnUserInput() {
	// Given
	materializedStatementResults := getResultsExample()
	s.initMockCalls(materializedStatementResults)
	s.interactiveOutputController.init()

	s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)
	s.resultFetcher.EXPECT().GetResults().Return(materializedStatementResults)

	// When
	result := s.interactiveOutputController.inputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

	// Then
	require.Nil(s.T(), result)
	require.True(s.T(), s.interactiveOutputController.isRowViewOpen)
	// last row should be selected
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.interactiveOutputController.tableView.GetSelectedRow())
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

func (s *InteractiveOutputControllerTestSuite) initMockCalls(materializedStatementResults *types.MaterializedStatementResults) {
	s.resultFetcher.EXPECT().SetAutoRefreshCallback(gomock.Any())
	s.resultFetcher.EXPECT().ToggleAutoRefresh()
	s.resultFetcher.EXPECT().GetResults().Return(materializedStatementResults)
	s.resultFetcher.EXPECT().IsTableMode().Return(materializedStatementResults.IsTableMode())
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Paused)
	s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)
}

func (s *InteractiveOutputControllerTestSuite) TestCloseRowViewOnUserInput() {
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
			s.interactiveOutputController.init()
			s.interactiveOutputController.isRowViewOpen = true

			// When
			result := s.interactiveOutputController.inputCapture(testCase.input)

			// Then
			require.Nil(s.T(), result)
			require.False(s.T(), s.interactiveOutputController.isRowViewOpen)
		})
	}
}

func (s *InteractiveOutputControllerTestSuite) TestNonSupportedUserInputInRowView() {
	// Given
	s.interactiveOutputController.isRowViewOpen = true

	// When
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := s.interactiveOutputController.inputCapture(input)

	// Then
	require.NotNil(s.T(), result)
	require.Equal(s.T(), input, result)
	require.True(s.T(), s.interactiveOutputController.isRowViewOpen)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysTableMode() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Failed)

	actual := s.interactiveOutputController.getTableTitle()

	require.Contains(s.T(), actual, "Table mode")
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysChangelogMode() {
	s.resultFetcher.EXPECT().IsTableMode().Return(false)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Failed)

	actual := s.interactiveOutputController.getTableTitle()

	require.Contains(s.T(), actual, "Changelog mode")
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysComplete() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Completed)

	actual := s.interactiveOutputController.getTableTitle()

	require.Contains(s.T(), actual, "(completed)")
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysFailed() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Failed)

	actual := s.interactiveOutputController.getTableTitle()

	require.Contains(s.T(), actual, "(auto refresh failed)")
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysPaused() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Paused)

	actual := s.interactiveOutputController.getTableTitle()

	require.Contains(s.T(), actual, "(auto refresh paused)")
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysRunning() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Running)

	actual := s.interactiveOutputController.getTableTitle()

	require.Contains(s.T(), actual, fmt.Sprintf("(auto refresh %vs)", results.DefaultRefreshInterval/1000))
}
