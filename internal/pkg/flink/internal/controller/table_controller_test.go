package controller

import (
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/gdamore/tcell/v2"
	"github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableControllerTestSuite struct {
	suite.Suite
	tableController *TableController
	appController   *mock.MockApplicationControllerInterface
	fetchController *mock.MockFetchControllerInterface
	dummyTViewApp   *tview.Application
}

func TestTableControllerTestSuite(t *testing.T) {
	suite.Run(t, new(TableControllerTestSuite))
}

func (s *TableControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.appController = mock.NewMockApplicationControllerInterface(ctrl)
	s.fetchController = mock.NewMockFetchControllerInterface(ctrl)
	s.dummyTViewApp = tview.NewApplication()
	s.tableController = NewTableController(components.CreateTable(), s.appController, s.fetchController, false).(*TableController)
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
			s.appController.EXPECT().SuspendOutputMode(gomock.Any())

			// When
			result := s.tableController.AppInputCapture(testCase.input)

			// Then
			require.Nil(s.T(), result)
		})
	}
}

func (s *TableControllerTestSuite) TestToggleTableModeOnUserInput() {
	// Given
	input := tcell.NewEventKey(tcell.KeyRune, 'M', tcell.ModNone)
	s.fetchController.EXPECT().ToggleTableMode()
	s.renderTableMockCalls()

	// When
	result := s.tableController.AppInputCapture(input)

	// Then
	require.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) renderTableMockCalls() {
	s.fetchController.EXPECT().IsTableMode().Return(true)
	s.fetchController.EXPECT().GetFetchState().Return(types.Paused)
	s.fetchController.EXPECT().GetMaxWidthPerColumn().Return([]int{})
	s.fetchController.EXPECT().GetHeaders().Return([]string{})
	s.fetchController.EXPECT().ForEach(gomock.Any())
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
	s.fetchController.EXPECT().GetResultsIterator(true)
	s.appController.EXPECT().TView().Return(s.dummyTViewApp)
}

func (s *TableControllerTestSuite) TestToggleRefreshResultsOnUserInput() {
	input := tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone)
	s.fetchController.EXPECT().ToggleAutoRefresh()
	s.renderTableMockCalls()

	result := s.tableController.AppInputCapture(input)

	require.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestNonSupportedUserInput() {
	// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
	// When we return the event, it's forwarded to tview
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)

	result := s.tableController.AppInputCapture(input)

	require.Equal(s.T(), input, result)
}

func (s *TableControllerTestSuite) TestOpenRowViewOnUserInput() {
	// Given
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
	s.fetchController.EXPECT().GetHeaders().Return(materializedStatementResults.GetHeaders())
	s.appController.EXPECT().TView().Return(s.dummyTViewApp).Times(2)

	// When
	result := s.tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

	// Then
	require.Nil(s.T(), result)
	require.True(s.T(), s.tableController.isRowViewOpen)
	require.Equal(s.T(), 10, s.tableController.selectedRowIdx) // 1-indexed: first row is at 1 and last row at 10
	// last row should be selected
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
}

func getMaterializedResultsExample() *types.MaterializedStatementResults {
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

func getStatementWithResultsExample() types.ProcessedStatement {
	statement := types.ProcessedStatement{
		StatementName: "example-statement",
		ResultSchema:  flinkgatewayv1alpha1.SqlV1alpha1ResultSchema{},
		StatementResults: &types.StatementResults{
			Headers: []string{"Count"},
			Rows:    []types.StatementResultRow{},
		},
	}
	for i := 0; i < 10; i++ {
		row := types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(i),
				},
			},
		}
		statement.StatementResults.Rows = append(statement.StatementResults.Rows, row)
	}
	return statement
}

func (s *TableControllerTestSuite) TestOpenRowViewWhenRowIsNilShouldNotPanic() {
	// Given
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})
	// move iterator to end so it becomes nil
	s.tableController.materializedStatementResultsIterator.Move(materializedStatementResults.Size() + 1)

	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
	s.fetchController.EXPECT().GetHeaders().Return(materializedStatementResults.GetHeaders())
	s.appController.EXPECT().TView().Return(s.dummyTViewApp).Times(2)

	// When
	result := s.tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

	// Then
	require.Nil(s.T(), result)
	require.True(s.T(), s.tableController.isRowViewOpen)
	require.Equal(s.T(), 10, s.tableController.selectedRowIdx) // 1-indexed: first row is at 1 and last row at 10
	require.Nil(s.T(), s.tableController.materializedStatementResultsIterator.Value())
}

func (s *TableControllerTestSuite) initMockCalls(materializedStatementResults *types.MaterializedStatementResults, fetchState types.FetchState) {
	s.fetchController.EXPECT().Init(gomock.Any())
	s.fetchController.EXPECT().SetAutoRefreshCallback(gomock.Any())
	s.fetchController.EXPECT().IsTableMode().Return(materializedStatementResults.IsTableMode())
	s.fetchController.EXPECT().GetFetchState().Return(fetchState)
	s.fetchController.EXPECT().GetMaxWidthPerColumn().Return(materializedStatementResults.GetMaxWidthPerColum())
	s.fetchController.EXPECT().GetHeaders().Return(materializedStatementResults.GetHeaders())
	s.fetchController.EXPECT().ForEach(gomock.Any()).Do(func(f func(rowIdx int, row *types.StatementResultRow)) { materializedStatementResults.ForEach(f) })
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false).Times(2)
	s.fetchController.EXPECT().GetResultsIterator(true).Return(materializedStatementResults.Iterator(true))
	s.appController.EXPECT().TView().Return(s.dummyTViewApp)
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
			s.tableController.isRowViewOpen = true

			// Expected mock calls
			s.appController.EXPECT().ShowTableView()
			s.appController.EXPECT().TView().Return(s.dummyTViewApp)

			// When
			result := s.tableController.AppInputCapture(testCase.input)

			// Then
			require.Nil(s.T(), result)
			require.False(s.T(), s.tableController.isRowViewOpen)
		})
	}
}

func (s *TableControllerTestSuite) TestSelectRowShouldDoNothingWhenRowToSelectSmallerThanOne() {
	materializedStatementResults := getMaterializedResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(-10, 0).Draw(t, "row to select")
		s.tableController.table.Select(rowToSelect, 0)

		// last row should be selected
		require.Equal(s.T(), materializedStatementResults.Size(), s.tableController.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
	})
}

func (s *TableControllerTestSuite) TestSelectRowShouldDoNothingWhenRowToSelectGreaterThanNumRows() {
	materializedStatementResults := getMaterializedResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(materializedStatementResults.Size()+1, materializedStatementResults.Size()+10).Draw(t, "row to select")
		s.tableController.table.Select(rowToSelect, 0)

		// last row should be selected
		require.Equal(s.T(), materializedStatementResults.Size(), s.tableController.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
	})
}

func (s *TableControllerTestSuite) TestSelectRowShouldDoNothingWhenAutoRefreshIsRunning() {
	materializedStatementResults := getMaterializedResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(1, materializedStatementResults.Size()).Draw(t, "row to select")
		s.fetchController.EXPECT().IsAutoRefreshRunning().Return(true)
		s.tableController.table.Select(rowToSelect, 0)

		// last row should be selected
		require.Equal(s.T(), materializedStatementResults.Size(), s.tableController.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
	})
}

func (s *TableControllerTestSuite) TestSelectRowShouldNotMoveIteratorOnFirstCall() {
	materializedStatementResults := getMaterializedResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	rapid.Check(s.T(), func(t *rapid.T) {
		// need to set this manually, because Init selected the last row already
		s.tableController.selectedRowIdx = -1
		s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)

		rowToSelect := rapid.IntRange(1, materializedStatementResults.Size()).Draw(t, "row to select")
		s.tableController.table.Select(rowToSelect, 0)

		require.Equal(s.T(), rowToSelect, s.tableController.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
	})
}

func (s *TableControllerTestSuite) TestSelectArbitraryRow() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(1, materializedStatementResults.Size()).Draw(t, "row to select")
		s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)
		s.tableController.table.Select(rowToSelect, 0)

		require.Equal(s.T(), rowToSelect, s.tableController.selectedRowIdx)
		expectedIterator := materializedStatementResults.Iterator(false)
		expectedIterator.Move(rowToSelect - 1)
		require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
	})
}

func (s *TableControllerTestSuite) TestNonSupportedUserInputInRowView() {
	// Given
	s.tableController.isRowViewOpen = true

	// When
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := s.tableController.AppInputCapture(input)

	// Then
	require.NotNil(s.T(), result)
	require.Equal(s.T(), input, result)
	require.True(s.T(), s.tableController.isRowViewOpen)
}

func (s *TableControllerTestSuite) TestFastScrollUp() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})
	s.tableController.numRowsToScroll = 9
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)

	result := s.tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone))

	require.Nil(s.T(), result)
	require.Equal(s.T(), 1, s.tableController.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
}

func (s *TableControllerTestSuite) TestFastScrollUpShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})
	s.tableController.numRowsToScroll = 20
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false)

	result := s.tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone))

	require.Nil(s.T(), result)
	require.Equal(s.T(), 1, s.tableController.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
}

func (s *TableControllerTestSuite) TestFastScrollDown() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})
	s.tableController.numRowsToScroll = 9
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false).Times(2)
	s.tableController.table.Select(1, 0)

	result := s.tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone))

	require.Nil(s.T(), result)
	require.Equal(s.T(), materializedStatementResults.Size(), s.tableController.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
}

func (s *TableControllerTestSuite) TestFastScrollDownShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})
	s.tableController.numRowsToScroll = 20
	s.fetchController.EXPECT().IsAutoRefreshRunning().Return(false).Times(2)
	s.tableController.table.Select(1, 0)

	result := s.tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone))

	require.Nil(s.T(), result)
	require.Equal(s.T(), materializedStatementResults.Size(), s.tableController.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableController.materializedStatementResultsIterator.Value())
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysTableMode() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysChangelogMode() {
	materializedStatementResults := getMaterializedResultsExample()
	materializedStatementResults.SetTableMode(false)
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysComplete() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Completed)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysFailed() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Failed)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysPaused() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Paused)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysRunning() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, types.Running)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysUnknown() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, 5)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableControllerTestSuite) TestTableTitleDisplaysPageSizeAndCacheSizeWithUnsafeTrace() {
	materializedStatementResults := getMaterializedResultsExample()
	s.initMockCalls(materializedStatementResults, 5)
	s.tableController.debug = true
	statement := getStatementWithResultsExample()
	s.fetchController.EXPECT().GetStatement().Return(statement)
	s.fetchController.EXPECT().GetMaterializedStatementResults().Return(materializedStatementResults).Times(3)
	s.tableController.Init(types.ProcessedStatement{})

	actual := s.tableController.table.GetTitle()

	cupaloy.SnapshotT(s.T(), actual)
}
