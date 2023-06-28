package components

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableViewTestSuite struct {
	suite.Suite
	tableView *TableView
}

func TestTableViewTestSuite(t *testing.T) {
	suite.Run(t, new(TableViewTestSuite))
}

func (s *TableViewTestSuite) SetupTest() {
	s.tableView = NewTableView()
}

func (s *TableViewTestSuite) TestFastScrollUp() {
	materializedStatementResults := getResultsExample()
	s.tableView.numRowsToScroll = materializedStatementResults.Size()
	s.tableView.RenderTable("title", materializedStatementResults, true)

	s.tableView.FastScrollUp()

	require.Equal(s.T(), 1, s.tableView.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
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

func (s *TableViewTestSuite) TestFastScrollUpShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getResultsExample()
	s.tableView.numRowsToScroll = 100
	s.tableView.RenderTable("title", materializedStatementResults, true)

	s.tableView.FastScrollUp()

	require.Equal(s.T(), 1, s.tableView.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestFastScrollDown() {
	materializedStatementResults := getResultsExample()
	s.tableView.numRowsToScroll = materializedStatementResults.Size()
	s.tableView.RenderTable("title", materializedStatementResults, true)
	s.tableView.table.Select(1, 0)

	s.tableView.FastScrollDown()

	require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestFastScrollDownShouldNotMoveOutFurtherThanMax() {
	materializedStatementResults := getResultsExample()
	s.tableView.numRowsToScroll = 100
	s.tableView.RenderTable("title", materializedStatementResults, true)
	s.tableView.table.Select(1, 0)

	s.tableView.FastScrollDown()

	require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.selectedRowIdx)
	expectedIterator := materializedStatementResults.Iterator(true)
	require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
}

func (s *TableViewTestSuite) TestSelectRowShouldDoNothingWhenRowToSelectSmallerThanOne() {
	materializedStatementResults := getResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.tableView.RenderTable("title", materializedStatementResults, true)

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(-10, 0).Draw(t, "row to select")
		s.tableView.table.Select(rowToSelect, 0)

		// last row should be selected
		require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestSelectRowShouldDoNothingWhenRowToSelectGreaterThanNumRows() {
	materializedStatementResults := getResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.tableView.RenderTable("title", materializedStatementResults, true)

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(materializedStatementResults.Size()+1, materializedStatementResults.Size()+10).Draw(t, "row to select")
		s.tableView.table.Select(rowToSelect, 0)

		// last row should be selected
		require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestSelectRowShouldDoNothingWhenRowSelectionDisabled() {
	materializedStatementResults := getResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.tableView.RenderTable("title", materializedStatementResults, true)

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(materializedStatementResults.Size()+1, materializedStatementResults.Size()+10).Draw(t, "row to select")
		s.tableView.isRowSelectionEnabled = false
		s.tableView.table.Select(rowToSelect, 0)

		// last row should be selected
		require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestSelectRowShouldNotMoveIteratorOnFirstCall() {
	materializedStatementResults := getResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.tableView.RenderTable("title", materializedStatementResults, true)

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(1, materializedStatementResults.Size()).Draw(t, "row to select")
		s.tableView.selectedRowIdx = -1 // manually reset it to -1 to trigger this case
		s.tableView.table.Select(rowToSelect, 0)

		// now the selectedIdx should be set correctly, but the iterator will still be at the last row
		require.Equal(s.T(), rowToSelect, s.tableView.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestRenderTableShouldResetIteratorAndSelectedIdx() {
	materializedStatementResults := getResultsExample()
	expectedIterator := materializedStatementResults.Iterator(true)
	s.tableView.RenderTable("title", materializedStatementResults, true)
	s.tableView.numRowsToScroll = materializedStatementResults.Size()

	rapid.Check(s.T(), func(t *rapid.T) {
		s.tableView.FastScrollUp()
		s.tableView.RenderTable("title", materializedStatementResults, true)

		require.Equal(s.T(), materializedStatementResults.Size(), s.tableView.selectedRowIdx)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestSelectArbitraryRow() {
	materializedStatementResults := getResultsExample()
	s.tableView.RenderTable("title", materializedStatementResults, true)

	rapid.Check(s.T(), func(t *rapid.T) {
		rowToSelect := rapid.IntRange(1, materializedStatementResults.Size()).Draw(t, "row to select")
		s.tableView.table.Select(rowToSelect, 0)

		require.Equal(s.T(), rowToSelect, s.tableView.selectedRowIdx)
		expectedIterator := materializedStatementResults.Iterator(false)
		expectedIterator.Move(rowToSelect - 1)
		require.Equal(s.T(), expectedIterator.Value(), s.tableView.GetSelectedRow())
	})
}

func (s *TableViewTestSuite) TestTableShouldSetTitle() {
	expected := "Test Title"
	materializedStatementResults := getResultsExample()
	s.tableView.RenderTable(expected, materializedStatementResults, true)

	actual := s.tableView.table.GetTitle()

	require.Equal(s.T(), expected, actual)
}
