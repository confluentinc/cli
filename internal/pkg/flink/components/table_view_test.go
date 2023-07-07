package components

import (
	"strconv"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableViewTestSuite struct {
	suite.Suite
	tableView     *TableView
	numRowsScroll int
}

func TestTableViewTestSuite(t *testing.T) {
	suite.Run(t, new(TableViewTestSuite))
}

func (s *TableViewTestSuite) SetupTest() {
	s.tableView = NewTableView().(*TableView)
	s.numRowsScroll = s.tableView.getNumRowsToScroll()
}

func (s *TableViewTestSuite) TestFastScrollUp() {
	materializedStatementResults := getResultsExample(s.numRowsScroll + 1)
	s.tableView.RenderTable("title", materializedStatementResults, true, nil, 0)

	s.tableView.FastScrollUp()

	require.Equal(s.T(), 1, s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func getResultsExample(numRows int) *types.MaterializedStatementResults {
	materializedStatementResults := types.NewMaterializedStatementResults([]string{"Count"}, 10)
	for i := 0; i < numRows; i++ {
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

func (s *TableViewTestSuite) TestFastScrollUpShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getResultsExample(s.numRowsScroll / 2)
	s.tableView.RenderTable("title", materializedStatementResults, true, nil, 0)

	s.tableView.FastScrollUp()

	require.Equal(s.T(), 1, s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestFastScrollDown() {
	materializedStatementResults := getResultsExample(s.numRowsScroll + 1)
	s.tableView.RenderTable("title", materializedStatementResults, true, nil, 0)
	s.tableView.table.Select(1, 0)

	s.tableView.FastScrollDown()

	require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestFastScrollDownShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getResultsExample(s.numRowsScroll / 2)
	s.tableView.RenderTable("title", materializedStatementResults, true, nil, 0)
	s.tableView.table.Select(1, 0)

	s.tableView.FastScrollDown()

	require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestSelectArbitraryRow() {
	materializedStatementResults := getResultsExample(10)
	s.tableView.RenderTable("title", materializedStatementResults, true, nil, 0)

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(1, materializedStatementResults.Size()).Draw(t, "row to select")
		s.tableView.table.Select(rowToSelect, 0)

		require.Equal(s.T(), rowToSelect, s.tableView.getSelectedRowIdx())
		expectedIterator := materializedStatementResults.Iterator(false)
		expectedIterator.Move(rowToSelect - 1)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestTableShouldSetTitle() {
	expected := "Test Title"
	materializedStatementResults := getResultsExample(10)
	s.tableView.RenderTable(expected, materializedStatementResults, true, nil, 0)

	actual := s.tableView.table.GetTitle()

	require.Equal(s.T(), expected, actual)
}

func (s *TableViewTestSuite) TestTableShortcutsWithAutoRefreshOff() {
	materializedStatementResults := getResultsExample(10)

	actual := s.tableView.getTableShortcuts(materializedStatementResults, false)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) TestTableShortcutsWithAutoRefreshOn() {
	materializedStatementResults := getResultsExample(10)

	actual := s.tableView.getTableShortcuts(materializedStatementResults, true)

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) TestTableInfoBarWithAutoRefreshOnAndNoTimestamp() {
	materializedStatementResults := getResultsExample(10)
	s.tableView.RenderTable("title", materializedStatementResults, true, nil, 0)

	actual := s.getInfoBarText()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) getInfoBarText() []string {
	view := s.tableView.infoBar.GetView()
	var items []string
	for i := 0; i < view.GetItemCount(); i++ {
		item := view.GetItem(i).(*tview.TextView)
		items = append(items, item.GetText(true))
	}
	return items
}

func (s *TableViewTestSuite) TestTableInfoBarWithAutoRefreshOffAndNoTimestamp() {
	materializedStatementResults := getResultsExample(10)
	s.tableView.RenderTable("title", materializedStatementResults, false, nil, 0)

	actual := s.getInfoBarText()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) TestTableInfoBarWithAutoRefreshOnAndValidTimestamp() {
	materializedStatementResults := getResultsExample(10)
	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.tableView.RenderTable("title", materializedStatementResults, true, &timestamp, 0)

	actual := s.getInfoBarText()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) TestTableInfoBarWithAutoRefreshOffAndValidTimestamp() {
	materializedStatementResults := getResultsExample(10)
	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.tableView.RenderTable("title", materializedStatementResults, false, &timestamp, 0)

	actual := s.getInfoBarText()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) TestTableInfoBarWhenTableTableHasNoContentAndAutoRefreshIsOff() {
	materializedStatementResults := getResultsExample(0)
	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.tableView.RenderTable("title", materializedStatementResults, false, &timestamp, 0)

	actual := s.getInfoBarText()

	cupaloy.SnapshotT(s.T(), actual)
}

func (s *TableViewTestSuite) TestTableInfoBarWhenTableTableHasNoContentAndAutoRefreshIsOn() {
	materializedStatementResults := getResultsExample(0)
	timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.tableView.RenderTable("title", materializedStatementResults, true, &timestamp, 0)

	actual := s.getInfoBarText()

	cupaloy.SnapshotT(s.T(), actual)
}
