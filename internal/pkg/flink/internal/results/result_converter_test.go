package results

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type ResultConverterTestSuite struct {
	suite.Suite
}

func TestResultConverterTestSuite(t *testing.T) {
	suite.Run(t, new(ResultConverterTestSuite))
}

func (s *ResultConverterTestSuite) TestConvertField() {
	rapid.Check(s.T(), func(t *rapid.T) {
		maxNestingDepth := rapid.IntRange(0, 5).Draw(t, "max nesting depth")
		dataType := generators.DataType(maxNestingDepth).Draw(t, "data type")
		field := generators.GetResultItemGeneratorForType(dataType).Draw(t, "a field")
		resultField := convertToInternalField(field, flinkgatewayv1alpha1.ColumnDetails{
			Name: "Test Column",
			Type: dataType,
		})
		require.NotNil(t, resultField)
		require.Equal(t, types.NewResultFieldType(dataType), resultField.GetType())
		if maxNestingDepth == 0 {
			require.IsType(t, types.AtomicStatementResultField{}, resultField)
		}
		require.Equal(t, field, resultField.ToSDKType())
	})
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForMissingDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, flinkgatewayv1alpha1.ColumnDetails{
		Name: "Test Column",
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.NULL, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, flinkgatewayv1alpha1.ColumnDetails{
		Name: "Test Column",
		Type: flinkgatewayv1alpha1.DataType{},
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.NULL, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsIfDataTypesDiffer() {
	varcharType := flinkgatewayv1alpha1.DataType{
		Nullable: false,
		Type:     "VARCHAR",
	}
	arrayType := flinkgatewayv1alpha1.DataType{
		Nullable:    false,
		Type:        "ARRAY",
		ElementType: &varcharType,
	}
	arrayField := generators.GetResultItemGeneratorForType(arrayType).Example()
	resultField := convertToInternalField(arrayField, flinkgatewayv1alpha1.ColumnDetails{
		Name: "Test Column",
		Type: varcharType,
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.NULL, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertResults() {
	rapid.Check(s.T(), func(t *rapid.T) {
		numColumns := rapid.IntRange(1, 10).Draw(t, "max nesting depth")
		results := generators.MockResults(numColumns, -1).Draw(t, "mock results")
		statementResults := results.StatementResults.Results.GetData()
		convertedResults, err := ConvertToInternalResults(statementResults, results.ResultSchema)
		require.NotNil(t, convertedResults)
		require.NoError(t, err)
		require.True(t, len(convertedResults.Headers) > 0)
		require.Equal(t, len(statementResults), len(convertedResults.Rows)) // row number should match
		for rowIdx, row := range convertedResults.Rows {
			expectedResultItem, ok := statementResults[rowIdx].(map[string]any)
			require.True(t, ok)
			op, ok := expectedResultItem["op"].(int32)
			require.True(t, ok)
			items, ok := expectedResultItem["row"].([]any)
			require.True(t, ok)

			require.Equal(t, types.StatementResultOperation(op), row.Operation)
			require.Equal(t, len(items), len(convertedResults.Headers)) // column number for this row should match
			for colIdx, field := range row.Fields {
				require.Equal(t, items[colIdx], field.ToSDKType()) // fields should match
			}
		}
	})
}

func (s *ResultConverterTestSuite) TestConvertResultsFailsWhenSchemaAndResultsDoNotMatch() {
	results := generators.MockResults(5, -1).Example()
	statementResults := results.StatementResults.GetResults()
	resultSchema := flinkgatewayv1alpha1.SqlV1alpha1ResultSchema{Columns: &[]flinkgatewayv1alpha1.ColumnDetails{}}
	internalResults, err := ConvertToInternalResults(statementResults.GetData(), resultSchema)
	require.Nil(s.T(), internalResults)
	require.Error(s.T(), err)
}
