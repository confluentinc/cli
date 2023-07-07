package controller

import (
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/gdamore/tcell/v2"
	"github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type InteractiveOutputControllerTestSuite struct {
	suite.Suite
	interactiveOutputController *InteractiveOutputController
	tableView                   *mock.MockTableViewInterface
	resultFetcher               *mock.MockResultFetcherInterface
	dummyTViewApp               *tview.Application
}

func TestInteractiveOutputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(InteractiveOutputControllerTestSuite))
}

func (s *InteractiveOutputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.tableView = mock.NewMockTableViewInterface(ctrl)
	s.resultFetcher = mock.NewMockResultFetcherInterface(ctrl)
	s.dummyTViewApp = tview.NewApplication()
	s.interactiveOutputController = NewInteractiveOutputController(s.tableView, s.resultFetcher, false).(*InteractiveOutputController)
}

func (s *InteractiveOutputControllerTestSuite) TestCloseTableViewOnUserInput() {
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Q", input: tcell.NewEventKey(tcell.KeyRune, rune(components.ExitTableViewShortcut[0]), tcell.ModNone)},
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
	input := tcell.NewEventKey(tcell.KeyRune, rune(components.ToggleTableModeShortcut[0]), tcell.ModNone)
	s.resultFetcher.EXPECT().ToggleTableMode()
	s.updateTableMockCalls(&types.MaterializedStatementResults{})

	result := s.interactiveOutputController.inputCapture(input)

	require.Nil(s.T(), result)
}

func (s *InteractiveOutputControllerTestSuite) updateTableMockCalls(materializedStatementResults *types.MaterializedStatementResults) {
	s.resultFetcher.EXPECT().IsTableMode().Return(true).Times(2)
	s.resultFetcher.EXPECT().GetFetchState().Return(types.Paused)
	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.resultFetcher.EXPECT().GetLastFetchTimestamp().Return(&timestamp)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(materializedStatementResults)
	s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)
	s.tableView.EXPECT().RenderTable(s.interactiveOutputController.getTableTitle(), materializedStatementResults, false, &timestamp, types.Paused)
	s.tableView.EXPECT().GetRoot().Return(tview.NewBox())
	s.tableView.EXPECT().GetFocusableElement().Return(tview.NewTable())
}

func (s *InteractiveOutputControllerTestSuite) TestToggleRefreshResultsOnUserInput() {
	s.initMockCalls(&types.MaterializedStatementResults{})
	s.interactiveOutputController.init()
	input := tcell.NewEventKey(tcell.KeyRune, rune(components.ToggleAutoRefreshShortcut[0]), tcell.ModNone)
	s.resultFetcher.EXPECT().ToggleAutoRefresh()
	s.updateTableMockCalls(&types.MaterializedStatementResults{})

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
	iterator := materializedStatementResults.Iterator(true)
	s.initMockCalls(materializedStatementResults)
	s.interactiveOutputController.init()

	s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(materializedStatementResults)
	s.tableView.EXPECT().GetSelectedRow().Return(iterator.Value())

	// When
	result := s.interactiveOutputController.inputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

	// Then
	require.Nil(s.T(), result)
	require.True(s.T(), s.interactiveOutputController.isRowViewOpen)
}

func getResultsExample() *types.MaterializedStatementResults {
	executedStatementWithResults := getStatementWithResultsExample()
	mat := types.NewMaterializedStatementResults(executedStatementWithResults.StatementResults.GetHeaders(), 10)
	mat.Append(executedStatementWithResults.StatementResults.GetRows()...)
	return &mat
}

func (s *InteractiveOutputControllerTestSuite) initMockCalls(materializedStatementResults *types.MaterializedStatementResults) {
	s.resultFetcher.EXPECT().SetAutoRefreshCallback(gomock.Any())
	s.resultFetcher.EXPECT().ToggleAutoRefresh()
	s.tableView.EXPECT().Init()
	s.updateTableMockCalls(materializedStatementResults)
}

func (s *InteractiveOutputControllerTestSuite) TestCloseRowViewOnUserInput() {
	// Given
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Q", input: tcell.NewEventKey(tcell.KeyRune, rune(components.ExitRowViewShortcut[0]), tcell.ModNone)},
		{name: "Test CtrlQ", input: tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)},
		{name: "Test Escape", input: tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			s.initMockCalls(&types.MaterializedStatementResults{})
			s.interactiveOutputController.init()
			s.interactiveOutputController.isRowViewOpen = true
			s.tableView.EXPECT().GetRoot().Return(tview.NewBox())
			s.tableView.EXPECT().GetFocusableElement().Return(tview.NewTable())

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

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysChangelogMode() {
	s.resultFetcher.EXPECT().IsTableMode().Return(false)

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysComplete() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysFailed() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysPaused() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysRunning() {
	s.resultFetcher.EXPECT().IsTableMode().Return(true)

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestTableTitleDisplaysPageSizeAndCacheSizeWithUnsafeTrace() {
	executedStatementWithResults := getStatementWithResultsExample()
	mat := types.NewMaterializedStatementResults(executedStatementWithResults.StatementResults.GetHeaders(), 10)
	mat.Append(executedStatementWithResults.StatementResults.GetRows()...)

	s.resultFetcher.EXPECT().IsTableMode().Return(true)
	s.resultFetcher.EXPECT().GetStatement().Return(executedStatementWithResults)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat).Times(3)
	s.interactiveOutputController.debug = true

	actual := s.interactiveOutputController.getTableTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *InteractiveOutputControllerTestSuite) TestArrowUpOrDownTogglesAutoRefreshWhenAutoRefreshIsRunning() {
	// Given
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Arrow Up", input: tcell.NewEventKey(tcell.KeyUp, rune(0), tcell.ModNone)},
		{name: "Test Arrow Down", input: tcell.NewEventKey(tcell.KeyDown, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			s.initMockCalls(&types.MaterializedStatementResults{})
			s.interactiveOutputController.init()
			s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(true)
			s.resultFetcher.EXPECT().ToggleAutoRefresh()
			s.updateTableMockCalls(&types.MaterializedStatementResults{})

			result := s.interactiveOutputController.inputCapture(testCase.input)

			require.Nil(t, result)
		})
	}
}

func (s *InteractiveOutputControllerTestSuite) TestArrowUpOrDownDoesNothingWhenAutoRefreshNotRunning() {
	// Given
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Arrow Up", input: tcell.NewEventKey(tcell.KeyUp, rune(0), tcell.ModNone)},
		{name: "Test Arrow Down", input: tcell.NewEventKey(tcell.KeyDown, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			s.initMockCalls(&types.MaterializedStatementResults{})
			s.interactiveOutputController.init()
			s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)

			result := s.interactiveOutputController.inputCapture(testCase.input)

			require.Equal(t, testCase.input, result)
		})
	}
}

func (s *InteractiveOutputControllerTestSuite) TestJumpUpOrDownTogglesAutoRefreshWhenAutoRefreshIsRunning() {
	// Given
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Jump Up", input: tcell.NewEventKey(tcell.KeyRune, rune(components.JumpUpShortcut[0]), tcell.ModNone)},
		{name: "Test Jump Down", input: tcell.NewEventKey(tcell.KeyRune, rune(components.JumpDownShortcut[0]), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			s.initMockCalls(&types.MaterializedStatementResults{})
			s.interactiveOutputController.init()
			s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(true)
			s.resultFetcher.EXPECT().ToggleAutoRefresh()
			s.updateTableMockCalls(&types.MaterializedStatementResults{})

			result := s.interactiveOutputController.inputCapture(testCase.input)

			require.Nil(t, result)
		})
	}
}

func (s *InteractiveOutputControllerTestSuite) TestJumpUpOrDownScrollsWhenAutoRefreshNotRunning() {
	// Given
	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Jump Up", input: tcell.NewEventKey(tcell.KeyRune, rune(components.JumpUpShortcut[0]), tcell.ModNone)},
		{name: "Test Jump Down", input: tcell.NewEventKey(tcell.KeyRune, rune(components.JumpDownShortcut[0]), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			s.initMockCalls(&types.MaterializedStatementResults{})
			s.interactiveOutputController.init()
			s.resultFetcher.EXPECT().IsAutoRefreshRunning().Return(false)
			if testCase.input.Rune() == rune(components.JumpUpShortcut[0]) {
				s.tableView.EXPECT().JumpUp()
			} else {
				s.tableView.EXPECT().JumpDown()
			}

			result := s.interactiveOutputController.inputCapture(testCase.input)

			require.Nil(t, result)
		})
	}
}
