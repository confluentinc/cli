package results

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/test/generators"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

const (
	maxNestingDepthLabel = "max nesting depth"
	testColumnName       = "Test Column"
)

type ResultConverterTestSuite struct {
	suite.Suite
}

func TestResultConverterTestSuite(t *testing.T) {
	suite.Run(t, new(ResultConverterTestSuite))
}

func (s *ResultConverterTestSuite) TestConvertField() {
	rapid.Check(s.T(), func(t *rapid.T) {
		maxNestingDepth := rapid.IntRange(0, 5).Draw(t, maxNestingDepthLabel)
		dataType := generators.DataType(maxNestingDepth).Draw(t, "data type")
		field := generators.GetResultItemGeneratorForType(dataType).Draw(t, "a field")
		resultField := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
			Name: testColumnName,
			Type: dataType,
		})
		require.NotNil(t, resultField)
		require.Equal(t, types.NewResultFieldType(dataType.GetType()), resultField.GetType())
		if maxNestingDepth == 0 {
			require.IsType(t, types.AtomicStatementResultField{}, resultField)
		}
		expected := normalizeExpected(field, dataType)
		require.Equal(t, expected, resultField.ToSDKType())
	})
}

// normalizeExpected converts a raw generated value `field` into the exact
// SDK shape produced by the StatementResultField.ToSDKType() implementations.
// - STRUCTURED_TYPE -> map[string]any
// - ROW            -> []any (positional)
// - ARRAY          -> []any (elements normalized)
// - MAP            -> []any of [key, value] pairs
// - MULTISET       -> []any of [value, count] pairs
// - atomic         -> as-is (note your Atomic ToSDKType returns string values)
func normalizeExpected(field any, dt flinkgatewayv1.DataType) any {
	typ := dt.GetType()
	switch typ {
	case "STRUCTURED_TYPE":
		// generator gives []any aligned with schema order; SDK expects map[name]value
		items, ok := field.([]any)
		fields := dt.GetFields()
		if !ok || fields == nil || len(items) != len(fields) {
			return field
		}
		out := make(map[string]any, len(items))
		for i, f := range fields {
			out[f.GetName()] = normalizeExpected(items[i], f.GetFieldType())
		}
		return out
	case "ROW":
		// stays positional []any
		items, ok := field.([]any)
		fields := dt.GetFields()
		if !ok || fields == nil || len(items) != len(fields) {
			return field
		}
		out := make([]any, len(items))
		for i, f := range fields {
			out[i] = normalizeExpected(items[i], f.GetFieldType())
		}
		return out
	case "ARRAY":
		elems, ok := field.([]any)
		if !ok {
			return field
		}
		elemType := dt.GetElementType()
		out := make([]any, len(elems))
		for i := range elems {
			if &elemType != nil {
				out[i] = normalizeExpected(elems[i], elemType)
			} else {
				out[i] = elems[i]
			}
		}
		return out
	case "MAP":
		// SDK shape is []any of [key, value] pairs
		entries, ok := field.([]any)
		if !ok {
			return field
		}
		keyT := dt.GetKeyType()
		valT := dt.GetValueType()
		out := make([]any, 0, len(entries))
		for _, e := range entries {
			pair, ok := e.([]any)
			if !ok || len(pair) != 2 {
				// best-effort: keep as-is
				out = append(out, e)
				continue
			}
			var k, v = pair[0], pair[1]
			if &keyT != nil {
				k = normalizeExpected(pair[0], keyT)
			}
			if &valT != nil {
				v = normalizeExpected(pair[1], valT)
			}
			out = append(out, []any{k, v})
		}
		return out
	case "MULTISET":
		// In our model, MULTISET is converted via MapStatementResultField with
		// keyType = elementType, valueType = INTEGER; and ToSDKType() returns []any of pairs.
		entries, ok := field.([]any)
		if !ok {
			return field
		}
		elemT := dt.GetElementType() // logical "key" type
		// value/count type is INTEGER in your converter, but Atomic returns strings, so keep as-is.
		out := make([]any, 0, len(entries))
		for _, e := range entries {
			// generator may produce either:
			// - the raw element (imply implicit count 1), or
			// - a pair-like []any{element, count}
			if pair, ok := e.([]any); ok && len(pair) >= 2 {
				val := pair[0]
				if &elemT != nil {
					val = normalizeExpected(pair[0], elemT)
				}
				// keep count as-is (often string "0"/"2" in your atomic model)
				cnt := pair[1]
				out = append(out, []any{val, cnt})
			} else {
				// implicit count=1; match SDK pair form
				val := e
				if &elemT != nil {
					val = normalizeExpected(e, elemT)
				}
				out = append(out, []any{val, "1"})
			}
		}
		return out
	default:
		// atomic or unhandled: return as-is (your Atomic ToSDKType returns string for numbers)
		return field
	}
}

func (s *ResultConverterTestSuite) TestConvertFieldOnPrem() {
	rapid.Check(s.T(), func(t *rapid.T) {
		maxNestingDepth := rapid.IntRange(0, 5).Draw(t, maxNestingDepthLabel)
		dataType := generators.DataTypeOnPrem(maxNestingDepth).Draw(t, "data type")
		field := generators.GetResultItemGeneratorForTypeOnPrem(dataType).Draw(t, "a field")
		resultField := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
			Name: testColumnName,
			Type: dataType,
		})
		require.NotNil(t, resultField)
		require.Equal(t, types.NewResultFieldType(dataType.GetType()), resultField.GetType())
		if maxNestingDepth == 0 {
			require.IsType(t, types.AtomicStatementResultField{}, resultField)
		}
		require.Equal(t, field, resultField.ToSDKType())
	})
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForMissingDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForMissingDataTypeOnPrem() {
	dataType := generators.DataTypeOnPrem(0).Example()
	field := generators.GetResultItemGeneratorForTypeOnPrem(dataType).Example()
	resultField := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
		Type: flinkgatewayv1.DataType{},
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataTypeOnPrem() {
	dataType := generators.DataTypeOnPrem(0).Example()
	field := generators.GetResultItemGeneratorForTypeOnPrem(dataType).Example()
	resultField := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
		Type: cmfsdk.DataType{},
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsIfDataTypesDiffer() {
	varcharType := flinkgatewayv1.DataType{
		Nullable: false,
		Type:     "VARCHAR",
	}
	arrayType := flinkgatewayv1.DataType{
		Nullable:    false,
		Type:        "ARRAY",
		ElementType: &varcharType,
	}
	arrayField := generators.GetResultItemGeneratorForType(arrayType).Example()
	resultField := convertToInternalField(arrayField, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
		Type: varcharType,
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsIfDataTypesDifferOnPrem() {
	varcharType := cmfsdk.DataType{
		Nullable: false,
		Type:     "VARCHAR",
	}
	arrayType := cmfsdk.DataType{
		Nullable:    false,
		Type:        "ARRAY",
		ElementType: &varcharType,
	}
	arrayField := generators.GetResultItemGeneratorForTypeOnPrem(arrayType).Example()
	resultField := convertToInternalFieldOnPrem(arrayField, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
		Type: varcharType,
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertResults() {
	rapid.Check(s.T(), func(t *rapid.T) {
		numColumns := rapid.IntRange(1, 10).Draw(t, maxNestingDepthLabel)
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
			op, ok := expectedResultItem["op"].(float64)
			require.True(t, ok)
			items, ok := expectedResultItem["row"].([]any)
			require.True(t, ok)

			require.Equal(t, types.StatementResultOperation(op), row.Operation)
			require.Equal(t, len(items), len(convertedResults.Headers)) // column number for this row should match
			for colIdx, field := range row.Fields {
				expected := items[colIdx]
				// normalize STRUCTURED_TYPE, MAP, MULTISET, ROW, ARRAY recursively
				expected = normalizeExpected(expected, (*results.ResultSchema.Columns)[colIdx].Type)
				require.Equal(t, expected, field.ToSDKType()) // fields should match
			}
		}
	})
}

func (s *ResultConverterTestSuite) TestConvertResultsOnPrem() {
	rapid.Check(s.T(), func(t *rapid.T) {
		numColumns := rapid.IntRange(1, 10).Draw(t, maxNestingDepthLabel)
		results := generators.MockResultsOnPrem(numColumns, -1).Draw(t, "mock results")
		statementResults := results.StatementResults.Results
		statementResultsData := results.StatementResults.Results.GetData()
		convertedResults, err := ConvertToInternalResultsOnPrem(statementResults, results.ResultSchema)
		require.NotNil(t, convertedResults)
		require.NoError(t, err)
		require.True(t, len(convertedResults.Headers) > 0)
		require.Equal(t, len(statementResultsData), len(convertedResults.Rows)) // row number should match
		for rowIdx, row := range convertedResults.Rows {
			expectedResultItem := statementResultsData[rowIdx]
			op, ok := expectedResultItem["op"].(float64)
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
	resultSchema := flinkgatewayv1.SqlV1ResultSchema{Columns: &[]flinkgatewayv1.ColumnDetails{}}
	internalResults, err := ConvertToInternalResults(statementResults.GetData(), resultSchema)
	require.Nil(s.T(), internalResults)
	require.Error(s.T(), err)
}

func (s *ResultConverterTestSuite) TestConvertResultsFailsWhenSchemaAndResultsDoNotMatchOnPrem() {
	results := generators.MockResultsOnPrem(5, -1).Example()
	statementResults := results.StatementResults.GetResults()
	resultSchema := cmfsdk.ResultSchema{Columns: []cmfsdk.ResultSchemaColumn{}}
	internalResults, err := ConvertToInternalResultsOnPrem(statementResults, resultSchema)
	require.Nil(s.T(), internalResults)
	require.Error(s.T(), err)
}
