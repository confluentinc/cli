package results

import (
	"strconv"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/v4/pkg/flink/test/generators"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

type MaterializedStatementResultsTestSuite struct {
	suite.Suite
}

func TestMaterializedStatementResultsTestSuite(t *testing.T) {
	suite.Run(t, new(MaterializedStatementResultsTestSuite))
}

func (s *MaterializedStatementResultsTestSuite) TestChangelogMode() {
	rapid.Check(s.T(), func(t *rapid.T) {
		// generate some results
		numColumns := rapid.IntRange(1, 10).Draw(t, "max nesting depth")
		results := generators.MockResults(numColumns, -1).Draw(t, "mock results")
		statementResults := results.StatementResults.Results.GetData()
		convertedResults, err := ConvertToInternalResults(statementResults, results.ResultSchema)
		require.NotNil(t, convertedResults)
		require.NoError(t, err)

		// test if in changelog mode all the rows are there and in the correct order
		materializedStatementResults := types.NewMaterializedStatementResults(convertedResults.GetHeaders(), 100, nil)
		materializedStatementResults.SetTableMode(false)
		materializedStatementResults.Append(convertedResults.GetRows()...)
		// in changelog mode we have an additional column "Operation"
		require.Equal(t, append([]string{"Operation"}, convertedResults.GetHeaders()...), materializedStatementResults.GetHeaders())
		require.Equal(t, len(convertedResults.GetRows()), materializedStatementResults.GetChangelogSize())
		iterator := materializedStatementResults.Iterator(false)
		for _, expectedRow := range convertedResults.GetRows() {
			actualRow := iterator.GetNext()
			operationField := types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: expectedRow.Operation.String(),
			}
			require.Equal(t, expectedRow.Operation, actualRow.Operation)
			// in changelog mode we have an additional column "Operation"
			require.Equal(t, append([]types.StatementResultField{operationField}, expectedRow.Fields...), actualRow.Fields)
		}
	})
}

func (s *MaterializedStatementResultsTestSuite) TestTableMode() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10000, nil)
	appendRow(&materializedStatementResults, types.Insert, types.Integer, "0")
	require.Equal(s.T(), headers, materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.GetTableSize())

	iterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "0",
			},
		},
	}, *iterator.GetNext())

	changelogSize := 1
	for count := 1; count <= 10; count++ {
		// remove previous row
		appendRow(&materializedStatementResults, types.UpdateBefore, types.Integer, strconv.Itoa(count-1))
		require.Equal(s.T(), 0, materializedStatementResults.GetTableSize())

		// add new row
		appendRow(&materializedStatementResults, types.UpdateAfter, types.Integer, strconv.Itoa(count))
		iterator = materializedStatementResults.Iterator(false)
		require.Equal(s.T(), 1, materializedStatementResults.GetTableSize())
		require.Equal(s.T(), types.StatementResultRow{
			Operation: types.UpdateAfter,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.Integer,
					Value: strconv.Itoa(count),
				},
			},
		}, *iterator.GetNext())
		changelogSize += 2
	}

	require.Equal(s.T(), changelogSize, materializedStatementResults.GetChangelogSize())
}

func (s *MaterializedStatementResultsTestSuite) TestKeyCountIncreases() {
	rapid.Check(s.T(), func(t *rapid.T) {
		keys := rapid.SliceOfNDistinct(rapid.IntRange(0, 11), 10, 10, rapid.ID[int]).Draw(t, "keys")

		// object under test
		headers := []string{"Key", "Count"}
		materializedStatementResults := types.NewMaterializedStatementResults(headers, 10000, nil)

		for _, key := range keys {
			appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(key), strconv.Itoa(1))
			appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(key), strconv.Itoa(1))
		}
		// check if the number of keys is correct
		require.Equal(t, 2*len(keys), materializedStatementResults.GetTableSize())
		require.Equal(t, 2*len(keys), materializedStatementResults.GetChangelogSize())

		for _, key := range keys {
			// Updates to group by are sent the following way
			// We get retraction for the previous groupby value
			// Then we get an update for the new groupby value
			appendRow(&materializedStatementResults, types.UpdateBefore, types.Integer, strconv.Itoa(key), strconv.Itoa(1))
			appendRow(&materializedStatementResults, types.UpdateAfter, types.Integer, strconv.Itoa(key), strconv.Itoa(0))
		}
		// check if the number of keys is correct
		require.Equal(t, 2*len(keys), materializedStatementResults.GetTableSize())
		require.Equal(t, 4*len(keys), materializedStatementResults.GetChangelogSize())

		// zeros
		table := materializedStatementResults.GetTable().ToSlice()
		zerosValues := lo.Filter(table, func(item types.StatementResultRow, _ int) bool {
			return item.Fields[1].ToString() == "0"
		})
		onesValues := lo.Filter(table, func(item types.StatementResultRow, _ int) bool {
			return item.Fields[1].ToString() == "1"
		})

		require.Equal(t, 10, len(zerosValues), "expected 10 zeros in table")
		require.Equal(t, 10, len(onesValues), "expected 10 ones in table")
	})
}

func (s *MaterializedStatementResultsTestSuite) TestChangelogIsCleanedUpWhenOverCapacity() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 3, nil)
	appendRow(&materializedStatementResults, types.Insert, types.Integer, "0")
	materializedStatementResults.SetTableMode(false)
	for count := 1; count <= 10; count++ {
		// remove previous row
		appendRow(&materializedStatementResults, types.UpdateBefore, types.Integer, strconv.Itoa(count-1))
		// add new row
		appendRow(&materializedStatementResults, types.UpdateAfter, types.Integer, strconv.Itoa(count))
	}

	require.Equal(s.T(), 3, materializedStatementResults.GetChangelogSize())
	require.Equal(s.T(), [][]string{{"+U", "9"}, {"-U", "9"}, {"+U", "10"}}, toSlice(&materializedStatementResults))
}

func (s *MaterializedStatementResultsTestSuite) TestTableIsCleanedUpWhenOverCapacity() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 3, nil)
	for count := 1; count <= 10; count++ {
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(count))
	}

	require.Equal(s.T(), 3, materializedStatementResults.GetTableSize())
	require.Equal(s.T(), [][]string{{"8"}, {"9"}, {"10"}}, toSlice(&materializedStatementResults))
}

func (s *MaterializedStatementResultsTestSuite) TestTableDoesNotCleanupEventsThatWereRemovedFromChangelog() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 2, nil)
	appendRow(&materializedStatementResults, types.Insert, types.Integer, "100")
	appendRow(&materializedStatementResults, types.Insert, types.Integer, "0")
	for count := 1; count <= 10; count++ {
		// remove previous row
		appendRow(&materializedStatementResults, types.UpdateBefore, types.Integer, strconv.Itoa(count-1))
		// add new row
		appendRow(&materializedStatementResults, types.UpdateAfter, types.Integer, strconv.Itoa(count))
	}
	// remove last row
	appendRow(&materializedStatementResults, types.UpdateBefore, types.Integer, "10")

	require.Equal(s.T(), 1, materializedStatementResults.GetTableSize())
	require.Equal(s.T(), [][]string{{"100"}}, toSlice(&materializedStatementResults))
}

func (s *MaterializedStatementResultsTestSuite) TestCleanup() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 1, nil)
	appendRow(&materializedStatementResults, types.Insert, types.Integer, "0")
	iterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "0",
			},
		},
	}, *iterator.GetNext())

	for count := 1; count <= 10; count++ {
		// add new row
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(count))
		iterator = materializedStatementResults.Iterator(false)
		require.Equal(s.T(), 1, materializedStatementResults.GetTableSize())
		require.Equal(s.T(), types.StatementResultRow{
			Operation: types.Insert,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.Integer,
					Value: strconv.Itoa(count),
				},
			},
		}, *iterator.GetNext())
	}

	require.Equal(s.T(), 1, materializedStatementResults.GetTableSize())
}

func (s *MaterializedStatementResultsTestSuite) TestCleanupCacheBehaviorWithUpsertKeys() {
	headers := []string{"ID", "Data"}
	// Upsert on ID, capacity set to 2
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 2, &[]int32{0})

	// R1
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "K1", "A")
	// R2
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "K2", "X")
	// Table: [ (K1,A), (K2,X) ], size 2. Cache[K1]:[ptrR1], Cache[K2]:[ptrR2]
	require.Equal(s.T(), 2, materializedStatementResults.GetTableSize())

	// R3
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "K1", "B")
	// Table before cleanup would be: [ (K1,A), (K2,X), (K1,B) ], size 3. Cache[K1]:[ptrR1, ptrR3], Cache[K2]:[ptrR2]
	// Cleanup is triggered (3 > 2).
	// removedElement is (K1,A). removedRowKey is "K1".
	// Cache for "K1" is [ptrR1, ptrR3]. listPtrs[0] (ptrR1) == removedElement (ptrR1).
	// Cache for "K1" becomes [ptrR3].
	// Table becomes: [ (K2,X), (K1,B) ], size 2.
	require.Equal(s.T(), 2, materializedStatementResults.GetTableSize(), "Table size after cleanup (1st)")
	require.Equal(s.T(), [][]string{
		{"K2", "X"}, // K2X was R2, K1A (R1) was removed
		{"K1", "B"}, // K1B was R3
	}, toSlice(&materializedStatementResults), "Table content after cleanup (1st)")

	// R4
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "K3", "Y")
	// Table before cleanup: [ (K2,X), (K1,B), (K3,Y) ], size 3. Cache[K1]:[ptrR3], Cache[K2]:[ptrR2], Cache[K3]:[ptrR4]
	// Cleanup is triggered (3 > 2).
	// removedElement is (K2,X). removedRowKey is "K2".
	// Cache for "K2" is [ptrR2]. listPtrs[0] (ptrR2) == removedElement (ptrR2).
	// Cache for "K2" is deleted.
	// Table becomes: [ (K1,B), (K3,Y) ], size 2.
	require.Equal(s.T(), 2, materializedStatementResults.GetTableSize(), "Table size after cleanup (2nd)")
	require.Equal(s.T(), [][]string{
		{"K1", "B"}, // K1B was R3
		{"K3", "Y"}, // K3Y was R4
	}, toSlice(&materializedStatementResults), "Table content after cleanup (2nd)")

	// Verify that a key whose only element was removed via cleanup is gone from cache (indirectly)
	// Add K2 again, then K4, K2 should be cleaned up.
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "K2", "Z") // R5. Table: [K1B, K3Y, K2Z]. Cleanup K1B. Table: [K3Y, K2Z]
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "K4", "W") // R6. Table: [K3Y, K2Z, K4W]. Cleanup K3Y. Table: [K2Z, K4W]
	require.Equal(s.T(), 2, materializedStatementResults.GetTableSize(), "Table size after more cleanups")
	require.Equal(s.T(), [][]string{
		{"K2", "Z"},
		{"K4", "W"},
	}, toSlice(&materializedStatementResults), "Table content after more cleanups")
}

func (s *MaterializedStatementResultsTestSuite) TestCleanupCacheBehaviorNoUpsertKeys() {
	headers := []string{"Data1", "Data2"}
	// No upsert columns (nil)
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)

	// Insert three identical rows. They will have the same cache key.
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "V1", "V2") // R1
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "V1", "V2") // R2
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "V1", "V2") // R3

	require.Equal(s.T(), 3, materializedStatementResults.GetTableSize(), "Table should have 3 identical rows")
	require.Equal(s.T(), [][]string{
		{"V1", "V2"},
		{"V1", "V2"},
		{"V1", "V2"},
	}, toSlice(&materializedStatementResults))

	// Delete operation should remove the last added identical row (R3)
	// In non-upsert mode, a Delete operation still uses the rowKey which is based on all fields.
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "V1", "V2")
	require.Equal(s.T(), 2, materializedStatementResults.GetTableSize(), "Table size after deleting one of three identical rows")
	require.Equal(s.T(), [][]string{
		{"V1", "V2"},
		{"V1", "V2"},
	}, toSlice(&materializedStatementResults), "Table should have two identical rows left")

	// Delete again, should remove R2
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "V1", "V2")
	require.Equal(s.T(), 1, materializedStatementResults.GetTableSize(), "Table size after deleting two of three identical rows")
	require.Equal(s.T(), [][]string{
		{"V1", "V2"},
	}, toSlice(&materializedStatementResults), "Table should have one row left")

	// Delete last one, should remove R1
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "V1", "V2")
	require.Equal(s.T(), 0, materializedStatementResults.GetTableSize(), "Table size after deleting all identical rows")
	require.Empty(s.T(), toSlice(&materializedStatementResults), "Table should be empty")
}

func (s *MaterializedStatementResultsTestSuite) TestOnlyAllowAppendWithSameSchema() {
	invalidHeaders := []string{"Count"}
	row := types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "0",
			},
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "0",
			},
		},
	}
	materializedStatementResults := types.NewMaterializedStatementResults(invalidHeaders, 1, nil)
	valuesInserted := materializedStatementResults.Append(row)
	require.False(s.T(), valuesInserted)
	require.Empty(s.T(), materializedStatementResults.GetTableSize())

	validHeaders := []string{"Count", "Count2"}
	materializedStatementResults = types.NewMaterializedStatementResults(validHeaders, 1, nil)
	valuesInserted = materializedStatementResults.Append(row)
	require.True(s.T(), valuesInserted)
	require.Equal(s.T(), 1, materializedStatementResults.GetTableSize())
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorForwardResetThenBackward() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)

	for i := 0; i < 10; i++ {
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(i))
	}
	require.Equal(s.T(), 10, materializedStatementResults.GetTableSize())

	iterator := materializedStatementResults.Iterator(false)
	count := 0
	for !iterator.HasReachedEnd() {
		row := iterator.GetNext()
		require.Equal(s.T(), &types.StatementResultRow{
			Operation: types.Insert,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.Integer,
					Value: strconv.Itoa(count),
				},
			},
		}, row)
		count++
	}
	require.Equal(s.T(), materializedStatementResults.GetTableSize(), count)

	iterator = materializedStatementResults.Iterator(true)
	for !iterator.HasReachedEnd() {
		row := iterator.GetPrev()
		count--
		require.Equal(s.T(), &types.StatementResultRow{
			Operation: types.Insert,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.Integer,
					Value: strconv.Itoa(count),
				},
			},
		}, row)
	}
	require.Equal(s.T(), 0, count)
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorForwardAndBackward() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)

	for i := 0; i < 10; i++ {
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(i))
	}
	require.Equal(s.T(), 10, materializedStatementResults.GetTableSize())

	iterator := materializedStatementResults.Iterator(false)
	row := iterator.GetNext()
	require.Equal(s.T(), "0", row.Fields[0].(types.AtomicStatementResultField).Value)
	row = iterator.GetNext()
	require.Equal(s.T(), "1", row.Fields[0].(types.AtomicStatementResultField).Value)
	row = iterator.GetPrev()
	require.Equal(s.T(), "2", row.Fields[0].(types.AtomicStatementResultField).Value)
	row = iterator.GetPrev()
	require.Equal(s.T(), "1", row.Fields[0].(types.AtomicStatementResultField).Value)
	row = iterator.GetPrev()
	require.Equal(s.T(), "0", row.Fields[0].(types.AtomicStatementResultField).Value)
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorMoveToEndThenMoveToStart() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)

	for i := 0; i < 10; i++ {
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(i))
	}
	require.Equal(s.T(), 10, materializedStatementResults.GetTableSize())

	iterator := materializedStatementResults.Iterator(false)
	iterator.Move(9)
	require.Equal(s.T(), &types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "9",
			},
		},
	}, iterator.Value())

	iterator.Move(-9)
	require.Equal(s.T(), &types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "0",
			},
		},
	}, iterator.Value())
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorMoveDoesNotWorkOnceEndWasReached() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)

	for i := 0; i < 10; i++ {
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(i))
	}
	require.Equal(s.T(), 10, materializedStatementResults.GetTableSize())

	var expected *types.StatementResultRow
	iterator := materializedStatementResults.Iterator(false)
	iterator.Move(10)
	require.Equal(s.T(), expected, iterator.Value())

	iterator.Move(-9)
	require.Equal(s.T(), expected, iterator.Value())

	iterator = materializedStatementResults.Iterator(true)
	iterator.Move(-9)
	require.Equal(s.T(), &types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Integer,
				Value: "0",
			},
		},
	}, iterator.Value())
}

func (s *MaterializedStatementResultsTestSuite) TestForEach() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)
	for i := 0; i < 10; i++ {
		appendRow(&materializedStatementResults, types.Insert, types.Integer, strconv.Itoa(i))
	}

	idx := 0
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		expectedRow := &types.StatementResultRow{
			Operation: types.Insert,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.Integer,
					Value: strconv.Itoa(idx),
				},
			},
		}

		require.Equal(s.T(), idx, rowIdx)
		require.Equal(s.T(), expectedRow, row)
		idx++
	})
}

func (s *MaterializedStatementResultsTestSuite) TestGetColumnWidths() {
	headers := []string{"1234", "12"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "12345", "1")
	require.Equal(s.T(), []int{5, 2}, materializedStatementResults.GetMaxWidthPerColumn())
}

func (s *MaterializedStatementResultsTestSuite) TestGetColumnWidthsWithMultiLineValues() {
	headers := []string{"1234", "12"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "12345 \n 123456789", "1\n123")
	require.Equal(s.T(), []int{10, 3}, materializedStatementResults.GetMaxWidthPerColumn())
}

func (s *MaterializedStatementResultsTestSuite) TestGetColumnWidthsChangelogMode() {
	headers := []string{"1234", "12"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)
	materializedStatementResults.SetTableMode(false)
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "12345", "1")
	require.Equal(s.T(), []int{len("Operation"), 5, 2}, materializedStatementResults.GetMaxWidthPerColumn())
}

func (s *MaterializedStatementResultsTestSuite) TestToggleTableMode() {
	headers := []string{"column1", "column2"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "12", "3")

	require.True(s.T(), materializedStatementResults.IsTableMode())
	require.Equal(s.T(), headers, materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 0, materializedStatementResults.Size())

	materializedStatementResults.SetTableMode(false)

	require.False(s.T(), materializedStatementResults.IsTableMode())
	require.Equal(s.T(), append([]string{"Operation"}, headers...), materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
}

func (s *MaterializedStatementResultsTestSuite) TestRowKeysShouldNotCollide() {
	headers := []string{"column1", "column2"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, nil)

	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "1", "23")
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "12", "3")
	appendRow(&materializedStatementResults, types.UpdateBefore, types.Varchar, "1", "23")
	require.Equal(s.T(), [][]string{{"12", "3"}}, toSlice(&materializedStatementResults))
}

func (s *MaterializedStatementResultsTestSuite) TestUpsertColumns() {
	headers := []string{"column1", "column2", "column3"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, &[]int32{0, 2})

	// insert 2 rows, update after should be treated the same as insert
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "key1", "1", "key2")
	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "2", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "1", "key2"},
		{"key1", "2", "key1"},
	}, toSlice(&materializedStatementResults))
	// update before should nothing
	appendRow(&materializedStatementResults, types.UpdateBefore, types.Varchar, "key1", "1", "key2")
	appendRow(&materializedStatementResults, types.UpdateBefore, types.Varchar, "key1", "2", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "1", "key2"},
		{"key1", "2", "key1"},
	}, toSlice(&materializedStatementResults))

	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "2", "key2")
	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "3", "key1")
	// another UpdateAfter will be ignored
	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "3", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "2", "key2"},
		{"key1", "3", "key1"},
	}, toSlice(&materializedStatementResults))

	// row will be deleted regardless of the value, only the keys are important
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "key1", "2", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "2", "key2"},
	}, toSlice(&materializedStatementResults))
}

func (s *MaterializedStatementResultsTestSuite) TestUpsertColumnsMultipleInserts() {
	headers := []string{"key", "val"}
	upsertCols := &[]int32{0} // Upsert on "key"

	// Part 1: Lifecycle with multiple entries for the same key (Inserts, UpdateAfter, Deletes)
	results := types.NewMaterializedStatementResults(headers, 10, upsertCols)

	// Inserts for keyA
	appendRow(&results, types.Insert, types.Varchar, "keyA", "val1") // R1
	appendRow(&results, types.Insert, types.Varchar, "keyA", "val2") // R2
	appendRow(&results, types.Insert, types.Varchar, "keyA", "val3") // R3
	require.Equal(s.T(), 3, results.GetTableSize(), "After 3 inserts for keyA")
	require.Equal(s.T(), [][]string{
		{"keyA", "val1"},
		{"keyA", "val2"},
		{"keyA", "val3"},
	}, toSlice(&results), "Content after 3 inserts for keyA")

	// UpdateAfter affects the last entry for keyA (R3)
	appendRow(&results, types.UpdateAfter, types.Varchar, "keyA", "val3_updated")
	require.Equal(s.T(), 3, results.GetTableSize(), "After UpdateAfter on keyA")
	require.Equal(s.T(), [][]string{
		{"keyA", "val1"},
		{"keyA", "val2"},
		{"keyA", "val3_updated"},
	}, toSlice(&results), "Content after UpdateAfter on keyA")

	// Deletes affect the last entry for keyA progressively
	appendRow(&results, types.Delete, types.Varchar, "keyA", "any_val") // Deletes R3 (val3_updated)
	require.Equal(s.T(), 2, results.GetTableSize(), "After 1st delete on keyA")
	require.Equal(s.T(), [][]string{
		{"keyA", "val1"},
		{"keyA", "val2"},
	}, toSlice(&results), "Content after 1st delete on keyA")

	appendRow(&results, types.Delete, types.Varchar, "keyA", "any_val") // Deletes R2 (val2)
	require.Equal(s.T(), 1, results.GetTableSize(), "After 2nd delete on keyA")
	require.Equal(s.T(), [][]string{
		{"keyA", "val1"},
	}, toSlice(&results), "Content after 2nd delete on keyA")

	appendRow(&results, types.Delete, types.Varchar, "keyA", "any_val") // Deletes R1 (val1)
	require.Equal(s.T(), 0, results.GetTableSize(), "After 3rd delete on keyA")
	require.Empty(s.T(), toSlice(&results), "Content after all deletes for keyA")

	// Part 2: UpdateAfter on a new key acts as Insert
	results = types.NewMaterializedStatementResults(headers, 10, upsertCols)
	appendRow(&results, types.Insert, types.Varchar, "keyB", "valB1") // Pre-existing data

	appendRow(&results, types.UpdateAfter, types.Varchar, "keyC", "valC1") // UA on new keyC
	require.Equal(s.T(), 2, results.GetTableSize(), "After UA on new keyC")
	require.Contains(s.T(), toSlice(&results), []string{"keyB", "valB1"})
	require.Contains(s.T(), toSlice(&results), []string{"keyC", "valC1"})

	// Part 3: UpdateBefore is ignored for table state in upsert mode
	results = types.NewMaterializedStatementResults(headers, 10, upsertCols)
	appendRow(&results, types.Insert, types.Varchar, "keyD", "valD1")
	appendRow(&results, types.Insert, types.Varchar, "keyD", "valD2")

	initialTableContents := toSlice(&results)
	initialTableSize := results.GetTableSize()
	initialChangelogSize := results.GetChangelogSize()

	appendRow(&results, types.UpdateBefore, types.Varchar, "keyD", "valD2") // UB for last element
	require.Equal(s.T(), initialTableSize, results.GetTableSize(), "Table size after UB (1st)")
	require.Equal(s.T(), initialTableContents, toSlice(&results), "Table content after UB (1st)")
	require.Equal(s.T(), initialChangelogSize+1, results.GetChangelogSize(), "Changelog size after UB (1st)")

	appendRow(&results, types.UpdateBefore, types.Varchar, "keyD", "valD1") // UB for first element
	require.Equal(s.T(), initialTableSize, results.GetTableSize(), "Table size after UB (2nd)")
	require.Equal(s.T(), initialTableContents, toSlice(&results), "Table content after UB (2nd)")
	require.Equal(s.T(), initialChangelogSize+2, results.GetChangelogSize(), "Changelog size after UB (2nd)")
}

func (s *MaterializedStatementResultsTestSuite) TestFallbackToEntireRowAsKeyIfUpsertColumnsFail() {
	headers := []string{"column1", "column2", "column3"}
	// choose an upsert column idx that's out of range
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10, &[]int32{0, 10})

	// insert 2 rows
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "key1", "1", "key2")
	appendRow(&materializedStatementResults, types.Insert, types.Varchar, "key1", "2", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "1", "key2"},
		{"key1", "2", "key1"},
	}, toSlice(&materializedStatementResults))
	// update before should delete both rows
	appendRow(&materializedStatementResults, types.UpdateBefore, types.Varchar, "key1", "1", "key2")
	appendRow(&materializedStatementResults, types.UpdateBefore, types.Varchar, "key1", "2", "key1")
	require.Empty(s.T(), toSlice(&materializedStatementResults))

	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "2", "key2")
	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "3", "key1")
	// another UpdateAfter will duplicate the row
	appendRow(&materializedStatementResults, types.UpdateAfter, types.Varchar, "key1", "3", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "2", "key2"},
		{"key1", "3", "key1"},
		{"key1", "3", "key1"},
	}, toSlice(&materializedStatementResults))

	// Delete with wrong value will be ignored, because we can't find the row
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "key1", "2", "key1")
	// Delete with correct value will only delete one of the duplicates
	appendRow(&materializedStatementResults, types.Delete, types.Varchar, "key1", "3", "key1")
	require.Equal(s.T(), [][]string{
		{"key1", "2", "key2"},
		{"key1", "3", "key1"},
	}, toSlice(&materializedStatementResults))
}

func appendRow(materializedStatementResults *types.MaterializedStatementResults, operation types.StatementResultOperation, typeOfValues types.StatementResultFieldType, rowValues ...string) {
	fields := []types.StatementResultField{}
	for _, value := range rowValues {
		fields = append(fields, types.AtomicStatementResultField{
			Type:  typeOfValues,
			Value: value,
		})
	}
	materializedStatementResults.Append(types.StatementResultRow{
		Operation: operation,
		Fields:    fields,
	})
}

func toSlice(materializedStatementResults *types.MaterializedStatementResults) [][]string {
	var actual [][]string
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		var fields []string
		for _, field := range row.GetFields() {
			fields = append(fields, field.ToString())
		}
		actual = append(actual, fields)
	})
	return actual
}
