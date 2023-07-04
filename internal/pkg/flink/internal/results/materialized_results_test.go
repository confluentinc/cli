package results

import (
	"strconv"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
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
		materializedStatementResults := types.NewMaterializedStatementResults(convertedResults.GetHeaders(), 100)
		materializedStatementResults.SetTableMode(false)
		materializedStatementResults.Append(convertedResults.GetRows()...)
		// in changelog mode we have an additional column "Operation"
		require.Equal(t, append([]string{"Operation"}, convertedResults.GetHeaders()...), materializedStatementResults.GetHeaders())
		require.Equal(t, len(convertedResults.GetRows()), materializedStatementResults.Size())
		iterator := materializedStatementResults.Iterator(false)
		for _, expectedRow := range convertedResults.GetRows() {
			actualRow := iterator.GetNext()
			operationField := types.AtomicStatementResultField{
				Type:  "VARCHAR",
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
	previousRow := types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
		},
	}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10000)
	materializedStatementResults.Append(previousRow)
	materializedStatementResults.SetTableMode(true)
	require.Equal(s.T(), headers, materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
	iterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), previousRow, *iterator.GetNext())
	changelogSize := 1
	for count := 1; count <= 10; count++ {
		// remove previous row
		previousRow.Operation = types.UPDATE_BEFORE
		materializedStatementResults.Append(previousRow)
		require.Equal(s.T(), 0, materializedStatementResults.Size())

		// add new row
		previousRow = types.StatementResultRow{
			Operation: types.UPDATE_AFTER,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(count),
				},
			},
		}
		materializedStatementResults.Append(previousRow)
		iterator = materializedStatementResults.Iterator(false)
		require.Equal(s.T(), 1, materializedStatementResults.Size())
		require.Equal(s.T(), previousRow, *iterator.GetNext())
		changelogSize += 2
	}

	materializedStatementResults.SetTableMode(false)
	require.Equal(s.T(), append([]string{"Operation"}, headers...), materializedStatementResults.GetHeaders())
	require.Equal(s.T(), changelogSize, materializedStatementResults.Size())
}

func (s *MaterializedStatementResultsTestSuite) TestKeyCountIncreases() {
	rapid.Check(s.T(), func(t *rapid.T) {
		keys := rapid.SliceOfNDistinct(rapid.IntRange(0, 11), 10, 10, rapid.ID[int]).Draw(t, "keys")

		// object under test
		headers := []string{"Key", "Count"}
		materializedStatementResults := types.NewMaterializedStatementResults(headers, 10000)

		for _, key := range keys {
			row := types.StatementResultRow{
				Operation: types.INSERT,
				Fields: []types.StatementResultField{
					types.AtomicStatementResultField{
						Type:  types.INTEGER,
						Value: strconv.Itoa(key),
					},
					types.AtomicStatementResultField{
						Type:  types.INTEGER,
						Value: strconv.Itoa(1),
					},
				},
			}
			materializedStatementResults.Append(row)
			materializedStatementResults.Append(row)
		}
		// check if the number of keys is correct
		require.Equal(t, 2*len(keys), materializedStatementResults.Size())

		for _, key := range keys {
			// Updates to group by are sent the following way
			// We get retraction for the previous groupby value
			// Then we get an update for the new groupby value
			row := types.StatementResultRow{
				Operation: types.UPDATE_BEFORE,
				Fields: []types.StatementResultField{
					types.AtomicStatementResultField{
						Type:  types.INTEGER,
						Value: strconv.Itoa(key),
					},
					types.AtomicStatementResultField{
						Type:  types.INTEGER,
						Value: strconv.Itoa(1),
					},
				},
			}

			materializedStatementResults.Append(row)

			row = types.StatementResultRow{
				Operation: types.UPDATE_AFTER,
				Fields: []types.StatementResultField{
					types.AtomicStatementResultField{
						Type:  types.INTEGER,
						Value: strconv.Itoa(key),
					},
					types.AtomicStatementResultField{
						Type:  types.INTEGER,
						Value: strconv.Itoa(0),
					},
				},
			}

			materializedStatementResults.Append(row)
		}
		// check if the number of keys is correct
		require.Equal(t, 2*len(keys), materializedStatementResults.Size())

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

func (s *MaterializedStatementResultsTestSuite) TestMaxCapacity() {
	headers := []string{"Count"}
	previousRow := types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
		},
	}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 1)
	materializedStatementResults.Append(previousRow)
	materializedStatementResults.SetTableMode(true)
	require.Equal(s.T(), headers, materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
	iterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), previousRow, *iterator.GetNext())
	for count := 1; count <= 10; count++ {
		// remove previous row
		previousRow.Operation = types.UPDATE_BEFORE
		materializedStatementResults.Append(previousRow)
		require.Equal(s.T(), 0, materializedStatementResults.Size())

		// add new row
		previousRow = types.StatementResultRow{
			Operation: types.UPDATE_AFTER,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(count),
				},
			},
		}
		materializedStatementResults.Append(previousRow)
		iterator = materializedStatementResults.Iterator(false)
		require.Equal(s.T(), 1, materializedStatementResults.Size())
		require.Equal(s.T(), previousRow, *iterator.GetNext())
	}

	materializedStatementResults.SetTableMode(false)
	require.Equal(s.T(), append([]string{"Operation"}, headers...), materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
}

func (s *MaterializedStatementResultsTestSuite) TestCleanup() {
	headers := []string{"Count"}
	row := types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
		},
	}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 1)
	materializedStatementResults.Append(row)
	materializedStatementResults.SetTableMode(true)
	iterator := materializedStatementResults.Iterator(false)
	require.Equal(s.T(), row, *iterator.GetNext())
	for count := 1; count <= 10; count++ {
		// add new row
		row = types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(count),
				},
			},
		}
		materializedStatementResults.Append(row)
		iterator = materializedStatementResults.Iterator(false)
		require.Equal(s.T(), 1, materializedStatementResults.Size())
		require.Equal(s.T(), row, *iterator.GetNext())
	}

	require.Equal(s.T(), 1, materializedStatementResults.Size())
}

func (s *MaterializedStatementResultsTestSuite) TestOnlyAllowAppendWithSameSchema() {
	invalidHeaders := []string{"Count"}
	row := types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
		},
	}
	materializedStatementResults := types.NewMaterializedStatementResults(invalidHeaders, 1)
	materializedStatementResults.SetTableMode(true)
	valuesInserted := materializedStatementResults.Append(row)
	require.False(s.T(), valuesInserted)
	require.Empty(s.T(), materializedStatementResults.Size())

	validHeaders := []string{"Count", "Count2"}
	materializedStatementResults = types.NewMaterializedStatementResults(validHeaders, 1)
	materializedStatementResults.SetTableMode(true)
	valuesInserted = materializedStatementResults.Append(row)
	require.True(s.T(), valuesInserted)
	require.Equal(s.T(), 1, materializedStatementResults.Size())
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorForwardResetThenBackward() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(true)

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
	require.Equal(s.T(), 10, materializedStatementResults.Size())

	iterator := materializedStatementResults.Iterator(false)
	count := 0
	for !iterator.HasReachedEnd() {
		row := iterator.GetNext()
		require.Equal(s.T(), &types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(count),
				},
			},
		}, row)
		count++
	}
	require.Equal(s.T(), materializedStatementResults.Size(), count)

	iterator = materializedStatementResults.Iterator(true)
	for !iterator.HasReachedEnd() {
		row := iterator.GetPrev()
		count--
		require.Equal(s.T(), &types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(count),
				},
			},
		}, row)
	}
	require.Equal(s.T(), 0, count)
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorForwardAndBackward() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(true)

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
	require.Equal(s.T(), 10, materializedStatementResults.Size())

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
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(true)

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
	require.Equal(s.T(), 10, materializedStatementResults.Size())

	iterator := materializedStatementResults.Iterator(false)
	iterator.Move(9)
	require.Equal(s.T(), &types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "9",
			},
		},
	}, iterator.Value())

	iterator.Move(-9)
	require.Equal(s.T(), &types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
		},
	}, iterator.Value())
}

func (s *MaterializedStatementResultsTestSuite) TestIteratorMoveDoesNotWorkOnceEndWasReached() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(true)

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
	require.Equal(s.T(), 10, materializedStatementResults.Size())

	var expected *types.StatementResultRow
	iterator := materializedStatementResults.Iterator(false)
	iterator.Move(10)
	require.Equal(s.T(), expected, iterator.Value())

	iterator.Move(-9)
	require.Equal(s.T(), expected, iterator.Value())

	iterator = materializedStatementResults.Iterator(true)
	iterator.Move(-9)
	require.Equal(s.T(), &types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "0",
			},
		},
	}, iterator.Value())
}

func (s *MaterializedStatementResultsTestSuite) TestForEach() {
	headers := []string{"Count"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(true)

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

	idx := 0
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		expectedRow := &types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
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
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(true)
	materializedStatementResults.Append(types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "12345",
			},
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "1",
			},
		},
	})

	require.Equal(s.T(), []int{5, 2}, materializedStatementResults.GetMaxWidthPerColumn())
}

func (s *MaterializedStatementResultsTestSuite) TestGetColumnWidthsChangelogMode() {
	headers := []string{"1234", "12"}
	materializedStatementResults := types.NewMaterializedStatementResults(headers, 10)
	materializedStatementResults.SetTableMode(false)
	materializedStatementResults.Append(types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "12345",
			},
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "1",
			},
		},
	})

	require.Equal(s.T(), []int{len("Operation"), 5, 2}, materializedStatementResults.GetMaxWidthPerColumn())
}

func (s *MaterializedStatementResultsTestSuite) TestToggleTableMode() {
	materializedStatementResults := types.NewMaterializedStatementResults([]string{"1234", "12"}, 10)

	require.True(s.T(), materializedStatementResults.IsTableMode())

	materializedStatementResults.SetTableMode(false)

	require.False(s.T(), materializedStatementResults.IsTableMode())
}
