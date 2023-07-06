package components

import (
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
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
	s.tableView.RenderTable("title", materializedStatementResults, true)

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
	s.tableView.RenderTable("title", materializedStatementResults, true)

	s.tableView.FastScrollUp()

	require.Equal(s.T(), 1, s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestFastScrollDown() {
	materializedStatementResults := getResultsExample(s.numRowsScroll + 1)
	s.tableView.RenderTable("title", materializedStatementResults, true)
	s.tableView.table.Select(1, 0)

	s.tableView.FastScrollDown()

	require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestFastScrollDownShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getResultsExample(s.numRowsScroll / 2)
	s.tableView.RenderTable("title", materializedStatementResults, true)
	s.tableView.table.Select(1, 0)

	s.tableView.FastScrollDown()

	require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.getSelectedRowIdx())
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestSelectArbitraryRow() {
	materializedStatementResults := getResultsExample(10)
	s.tableView.RenderTable("title", materializedStatementResults, true)

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
	s.tableView.RenderTable(expected, materializedStatementResults, true)

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
