package results

import (
	"encoding/json"
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
		resultField, err := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
			Name: testColumnName,
			Type: dataType,
		})
		require.NoError(t, err)
		require.NotNil(t, resultField)
		expectedType, err := types.NewResultFieldType(dataType.GetType())
		require.NoError(t, err)
		require.Equal(t, expectedType, resultField.GetType())
		if maxNestingDepth == 0 {
			require.IsType(t, types.AtomicStatementResultField{}, resultField)
		}
		expected := normalizeExpected(field, dataType)
		require.Equal(t, expected, resultField.ToSDKType())
	})
}

// normalizeExpected transforms a raw, generator-produced value into the same
// shape returned by StatementResultField.ToSDKType(). This ensures property-
// based tests can compare generator output against conversion results.
//
// It recursively normalizes composite data types (ARRAY, MAP, MULTISET, ROW,
// STRUCTURED_TYPE) so that nested fields, key/value pairs, and element counts
// match the SDK representation. Atomic values are returned as-is.
func normalizeExpected(result any, dataType flinkgatewayv1.DataType) any {
	switch dataType.GetType() {
	case "STRUCTURED_TYPE":
		return normalizeStructuredType(result, dataType)
	case "ROW":
		return normalizeRow(result, dataType)
	case "ARRAY":
		return normalizeArray(result, dataType)
	case "MAP":
		return normalizeMap(result, dataType)
	case "MULTISET":
		return normalizeMultiSet(result, dataType)
	default:
		// atomic or unhandled: return as-is (our Atomic ToSDKType returns string for numbers)
		return result
	}
}

// normalizeRow handles ROW types by normalizing each field
// positionally according to its schema definition.
func normalizeRow(result any, dataType flinkgatewayv1.DataType) any {
	rawResults, ok := result.([]any)
	fields := dataType.GetFields()
	if !ok || fields == nil || len(rawResults) != len(fields) {
		return result
	}
	normalized := make([]any, len(rawResults))
	for i, field := range fields {
		normalized[i] = normalizeExpected(rawResults[i], field.GetFieldType())
	}
	return normalized
}

// normalizeStructuredType handles STRUCTURED_TYPE values by
// converting them into a map keyed by field names.
func normalizeStructuredType(result any, dataType flinkgatewayv1.DataType) any {
	rawResults, ok := result.([]any)
	fields := dataType.GetFields()
	if !ok || fields == nil || len(rawResults) != len(fields) {
		return result
	}
	normalized := make(map[string]any, len(rawResults))
	for i, field := range fields {
		normalized[field.GetName()] = normalizeExpected(rawResults[i], field.GetFieldType())
	}
	return normalized
}

// normalizeMap handles MAP values by normalizing them into
// slices of [key, value] pairs with recursive normalization.
func normalizeMap(field any, dataType flinkgatewayv1.DataType) any {
	rawEntries, ok := field.([]any)
	if !ok {
		return field
	}
	keyT := dataType.GetKeyType()
	valT := dataType.GetValueType()

	normalized := make([]any, 0, len(rawEntries))
	for _, rawEntry := range rawEntries {
		pair, ok := rawEntry.([]any)
		if !ok || len(pair) != 2 {
			normalized = append(normalized, rawEntry)
			continue
		}
		key, value := pair[0], pair[1]
		if keyT.GetType() != "" {
			key = normalizeExpected(key, keyT)
		}
		if valT.GetType() != "" {
			value = normalizeExpected(value, valT)
		}
		normalized = append(normalized, []any{key, value})
	}
	return normalized
}

// normalizeArray handles ARRAY values by normalizing each
// element recursively if the element type is set.
func normalizeArray(result any, dataType flinkgatewayv1.DataType) any {
	elems, ok := result.([]any)
	if !ok {
		return result
	}
	elemType := dataType.GetElementType()
	normalized := make([]any, len(elems))
	for i := range elems {
		if elemType.GetType() != "" {
			normalized[i] = normalizeExpected(elems[i], elemType)
		} else {
			normalized[i] = elems[i]
		}
	}
	return normalized
}

// normalizeMultiSet handles MULTISET values by normalizing them
// into slices of [element, count] pairs, where element is normalized
// recursively and count is preserved as-is.
func normalizeMultiSet(result any, dataType flinkgatewayv1.DataType) any {
	entries, ok := result.([]any)
	if !ok {
		return result
	}
	elemType := dataType.GetElementType()
	normalized := make([]any, 0, len(entries))
	for _, entry := range entries {
		pair, ok := entry.([]any)
		if !ok || len(pair) != 2 {
			normalized = append(normalized, entry)
			continue
		}
		value := pair[0]
		if elemType.GetType() != "" {
			value = normalizeExpected(pair[0], elemType)
		}
		count := pair[1]
		normalized = append(normalized, []any{value, count})
	}
	return normalized
}

func (s *ResultConverterTestSuite) TestConvertVariantFieldNil() {
	resultField := convertToInternalField(nil, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
		Type: flinkgatewayv1.DataType{Type: string(types.Variant)},
	})
	require.Equal(s.T(), types.Null, resultField.GetType())
}

func (s *ResultConverterTestSuite) TestConvertVariantNestedInArray() {
	variantType := flinkgatewayv1.DataType{Type: string(types.Variant)}
	column := flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
		Type: flinkgatewayv1.DataType{Type: string(types.Array), ElementType: &variantType},
	}
	var raw any
	require.NoError(s.T(), json.Unmarshal([]byte(`[[1,[["a",[11,"x"]]]],[0]]`), &raw))
	resultField := convertToInternalField(raw, column)
	require.Equal(s.T(), types.Array, resultField.GetType())
	require.Equal(s.T(), `[{"a":"x"}, null]`, resultField.ToString())
}

func (s *ResultConverterTestSuite) TestConvertVariantField() {
	rapid.Check(s.T(), func(t *rapid.T) {
		dataType := generators.VariantDataType().Draw(t, "data type")
		field := generators.VariantResultItem().Draw(t, "a field")
		resultField := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
			Name: testColumnName,
			Type: dataType,
		})
		require.NotNil(t, resultField)
		require.Equal(t, types.Variant, resultField.GetType())
		require.IsType(t, types.VariantStatementResultField{}, resultField)
		// A VARIANT cell always renders as valid JSON.
		var parsed any
		require.NoError(t, json.Unmarshal([]byte(resultField.ToString()), &parsed))
	})
}

func (s *ResultConverterTestSuite) TestConvertVariantFieldOnPrem() {
	rapid.Check(s.T(), func(t *rapid.T) {
		dataType := generators.VariantDataTypeOnPrem().Draw(t, "data type")
		field := generators.VariantResultItem().Draw(t, "a field")
		resultField := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
			Name: testColumnName,
			Type: dataType,
		})
		require.NotNil(t, resultField)
		require.Equal(t, types.Variant, resultField.GetType())
		require.IsType(t, types.VariantStatementResultField{}, resultField)
		var parsed any
		require.NoError(t, json.Unmarshal([]byte(resultField.ToString()), &parsed))
	})
}

func (s *ResultConverterTestSuite) TestConvertFieldOnPrem() {
	rapid.Check(s.T(), func(t *rapid.T) {
		maxNestingDepth := rapid.IntRange(0, 5).Draw(t, maxNestingDepthLabel)
		dataType := generators.DataTypeOnPrem(maxNestingDepth).Draw(t, "data type")
		field := generators.GetResultItemGeneratorForTypeOnPrem(dataType).Draw(t, "a field")
		resultField, err := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
			Name: testColumnName,
			Type: dataType,
		})
		require.NoError(t, err)
		require.NotNil(t, resultField)
		expectedType, err := types.NewResultFieldType(dataType.GetType())
		require.NoError(t, err)
		require.Equal(t, expectedType, resultField.GetType())
		if maxNestingDepth == 0 {
			require.IsType(t, types.AtomicStatementResultField{}, resultField)
		}
		require.Equal(t, field, resultField.ToSDKType())
	})
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForMissingDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	_, err := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
	})
	require.ErrorContains(s.T(), err, "unsupported result field type")
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForMissingDataTypeOnPrem() {
	dataType := generators.DataTypeOnPrem(0).Example()
	field := generators.GetResultItemGeneratorForTypeOnPrem(dataType).Example()
	_, err := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
	})
	require.ErrorContains(s.T(), err, "unsupported result field type")
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	_, err := convertToInternalField(field, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
		Type: flinkgatewayv1.DataType{},
	})
	require.ErrorContains(s.T(), err, "unsupported result field type")
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataTypeOnPrem() {
	dataType := generators.DataTypeOnPrem(0).Example()
	field := generators.GetResultItemGeneratorForTypeOnPrem(dataType).Example()
	_, err := convertToInternalFieldOnPrem(field, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
		Type: cmfsdk.DataType{},
	})
	require.ErrorContains(s.T(), err, "unsupported result field type")
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
	resultField, err := convertToInternalField(arrayField, flinkgatewayv1.ColumnDetails{
		Name: testColumnName,
		Type: varcharType,
	})
	require.NoError(s.T(), err)
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
	resultField, err := convertToInternalFieldOnPrem(arrayField, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
		Type: varcharType,
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.Null, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

// TestConvertFieldStructuredTypeOnPrem guards against GetConverterForTypeOnPrem
// silently regressing to null on STRUCTURED_TYPE, as it did before this test
// existed: it had no case for it (unlike the cloud GetConverterForType), and the
// property-based tests above never caught this because the on-prem generators
// never generate a STRUCTURED_TYPE column to begin with.
func (s *ResultConverterTestSuite) TestConvertFieldStructuredTypeOnPrem() {
	nameType := cmfsdk.DataType{Nullable: false, Type: "VARCHAR"}
	structuredType := cmfsdk.DataType{
		Nullable: false,
		Type:     "STRUCTURED_TYPE",
		Fields:   &[]cmfsdk.DataTypeField{*cmfsdk.NewDataTypeField("name", nameType)},
	}

	resultField, err := convertToInternalFieldOnPrem([]any{"Jim"}, cmfsdk.ResultSchemaColumn{
		Name: testColumnName,
		Type: structuredType,
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.StructuredType, resultField.GetType())
	require.IsType(s.T(), types.StructuredTypeStatementResultField{}, resultField)
	require.Equal(s.T(), map[string]any{"name": "Jim"}, resultField.ToSDKType())
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
