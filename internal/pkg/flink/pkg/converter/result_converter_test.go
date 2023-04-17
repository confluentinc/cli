package converter

import (
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"
	"testing"
)

type ResultConverterTestSuite struct {
	suite.Suite
}

func TestResultConverterTestSuite(t *testing.T) {
	suite.Run(t, new(ResultConverterTestSuite))
}

func (s *ResultConverterTestSuite) SetupSuite() {
}

func (s *ResultConverterTestSuite) TestConvertField() {
	rapid.Check(s.T(), func(t *rapid.T) {
		maxNestingDepth := rapid.IntRange(0, 5).Draw(t, "max nesting depth")
		dataType := types.DataType(maxNestingDepth).Draw(t, "data type")
		field := types.GetResultItemGeneratorForType(dataType).Draw(t, "a field")
		resultField := convertToInternalField(field, v1.ColumnDetails{
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
	dataType := types.DataType(0).Example()
	field := types.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, v1.ColumnDetails{
		Name: "Test Column",
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.NULL, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataType() {
	dataType := types.DataType(0).Example()
	field := types.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, v1.ColumnDetails{
		Name: "Test Column",
		Type: v1.DataType{},
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.NULL, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsIfDataTypesDiffer() {
	varcharType := v1.VarcharTypeAsDataType(&v1.VarcharType{
		Nullable: false,
		Type:     "VARCHAR",
	})
	arrayType := v1.ArrayTypeAsDataType(&v1.ArrayType{
		Nullable:         false,
		Type:             "ARRAY",
		ArrayElementType: varcharType,
	})
	arrayField := types.GetResultItemGeneratorForType(arrayType).Example()
	resultField := convertToInternalField(arrayField, v1.ColumnDetails{
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
		results := types.MockResults(numColumns).Draw(t, "mock results")
		statementResults := results.StatementResults.Results.GetData()
		convertedResults, err := ConvertToInternalResults(statementResults, results.ResultSchema)
		require.NotNil(s.T(), convertedResults)
		require.NoError(s.T(), err)
		require.True(s.T(), len(convertedResults) > 0)
		require.Equal(s.T(), len(statementResults), len(convertedResults[0].Fields)) // row number should match
		for rowIdx := range convertedResults[0].Fields {
			rowItem := statementResults[rowIdx].GetRow()
			items := rowItem.Items
			require.Equal(s.T(), len(items), len(convertedResults)) // column number for this row should match
			for colIdx, column := range convertedResults {
				require.Equal(s.T(), items[colIdx], column.Fields[rowIdx].ToSDKType()) // fields should match
			}
		}
	})
}

func (s *ResultConverterTestSuite) TestConvertResultsFailsWhenSchemaAndResultsDoNotMatch() {
	results := types.MockResults(5).Example()
	statementResults := results.StatementResults.GetResults()
	resultSchema := v1.SqlV1alpha1ResultSchema{Columns: &[]v1.ColumnDetails{}}
	internalResults, err := ConvertToInternalResults(statementResults.GetData(), resultSchema)
	require.Nil(s.T(), internalResults)
	require.Error(s.T(), err)
}
