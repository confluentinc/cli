package results

import (
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/confluentinc/flink-sql-client/test/generators"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"
	"strconv"
	"testing"
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
		materializedStatementResults := NewMaterializedStatementResults(convertedResults.GetHeaders(), 100)
		materializedStatementResults.SetTableMode(false)
		materializedStatementResults.AppendAll(convertedResults.GetRows())
		// in changelog mode we have an additional column "Operation"
		require.Equal(t, append([]string{"Operation"}, convertedResults.GetHeaders()...), materializedStatementResults.GetHeaders())
		require.Equal(t, len(convertedResults.GetRows()), materializedStatementResults.Size())
		iterator := materializedStatementResults.Iterator()
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
	materializedStatementResults := NewMaterializedStatementResults(headers, 100)
	materializedStatementResults.Append(previousRow)
	materializedStatementResults.SetTableMode(true)
	require.Equal(s.T(), headers, materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
	iterator := materializedStatementResults.Iterator()
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
		iterator = materializedStatementResults.Iterator()
		require.Equal(s.T(), 1, materializedStatementResults.Size())
		require.Equal(s.T(), previousRow, *iterator.GetNext())
		changelogSize += 2
	}

	materializedStatementResults.SetTableMode(false)
	require.Equal(s.T(), append([]string{"Operation"}, headers...), materializedStatementResults.GetHeaders())
	require.Equal(s.T(), changelogSize, materializedStatementResults.Size())
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
	materializedStatementResults := NewMaterializedStatementResults(headers, 1)
	materializedStatementResults.Append(previousRow)
	materializedStatementResults.SetTableMode(true)
	require.Equal(s.T(), headers, materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
	iterator := materializedStatementResults.Iterator()
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
		iterator = materializedStatementResults.Iterator()
		require.Equal(s.T(), 1, materializedStatementResults.Size())
		require.Equal(s.T(), previousRow, *iterator.GetNext())
	}

	materializedStatementResults.SetTableMode(false)
	require.Equal(s.T(), append([]string{"Operation"}, headers...), materializedStatementResults.GetHeaders())
	require.Equal(s.T(), 1, materializedStatementResults.Size())
}
