package results

import (
	"fmt"
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
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
		dataType := generators.DataType(maxNestingDepth).Draw(t, "data type")
		field := generators.GetResultItemGeneratorForType(dataType).Draw(t, "a field")
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
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
	resultField := convertToInternalField(field, v1.ColumnDetails{
		Name: "Test Column",
	})
	require.NotNil(s.T(), resultField)
	require.Equal(s.T(), types.NULL, resultField.GetType())
	require.IsType(s.T(), types.AtomicStatementResultField{}, resultField)
}

func (s *ResultConverterTestSuite) TestConvertFieldFailsForEmptyDataType() {
	dataType := generators.DataType(0).Example()
	field := generators.GetResultItemGeneratorForType(dataType).Example()
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
	arrayField := generators.GetResultItemGeneratorForType(arrayType).Example()
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
		results := generators.MockResults(numColumns, -1).Draw(t, "mock results")
		statementResults := results.StatementResults.Results.GetData()
		convertedResults, err := ConvertToInternalResults(statementResults, results.ResultSchema)
		require.NotNil(t, convertedResults)
		require.NoError(t, err)
		require.True(t, len(convertedResults.Headers) > 0)
		require.Equal(t, len(statementResults), len(convertedResults.Rows)) // row number should match
		for rowIdx, row := range convertedResults.Rows {
			op := statementResults[rowIdx].GetOp()
			rowItem := statementResults[rowIdx].GetRow()
			items := rowItem.Items
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
	resultSchema := v1.SqlV1alpha1ResultSchema{Columns: &[]v1.ColumnDetails{}}
	internalResults, err := ConvertToInternalResults(statementResults.GetData(), resultSchema)
	require.Nil(s.T(), internalResults)
	require.Error(s.T(), err)
}

func (s *ResultConverterTestSuite) TestFormatAtomicField() {
	rapid.Check(s.T(), func(t *rapid.T) {
		atomicDataType := generators.AtomicDataType().Draw(t, "atomic data type")
		atomicField := generators.GetResultItemGeneratorForType(atomicDataType).Draw(t, "atomic result field")
		convertedField := convertToInternalField(atomicField, v1.ColumnDetails{
			Name: "Test_Column",
			Type: atomicDataType,
		})

		val := "NULL"
		if types.NewResultFieldType(atomicDataType) != types.NULL {
			val = string(*atomicField.SqlV1alpha1ResultItemString)
		}

		require.Equal(s.T(), val, convertedField.Format(nil))
		maxDisplayableCharCount := rapid.IntRange(-3, 20).Draw(t, "max displayable chars")
		if len(val) > maxDisplayableCharCount {
			if maxDisplayableCharCount <= 3 {
				require.Equal(s.T(), "...", convertedField.Format(&types.FormatterOptions{MaxCharCountToDisplay: maxDisplayableCharCount}))
			} else {
				require.Equal(s.T(), val[:maxDisplayableCharCount-3]+"...", convertedField.Format(&types.FormatterOptions{MaxCharCountToDisplay: maxDisplayableCharCount}))
			}
		} else {
			require.Equal(s.T(), val, convertedField.Format(&types.FormatterOptions{MaxCharCountToDisplay: maxDisplayableCharCount}))
		}
	})
}

func (s *ResultConverterTestSuite) TestFormatArrayField() {
	arrayField := types.ArrayStatementResultField{
		Type:        types.ARRAY,
		ElementType: types.VARCHAR,
		Values: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "Test",
			},
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "Hello",
			},
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "World",
			},
		},
	}

	testCases := []struct {
		expected string
		options  *types.FormatterOptions
	}{
		{
			expected: "[Test, Hello, World]",
			options:  nil,
		},
		{
			expected: "...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 0},
		},
		{
			expected: "[...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 4},
		},
		{
			expected: "[Test, ...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 10},
		},
		{
			expected: "[Test, Hello, Wo...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 19},
		},
		{
			expected: "[Test, Hello, World]",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 20},
		},
	}

	for idx, testCase := range testCases {
		fmt.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := arrayField.Format(testCase.options)
		if testCase.options.GetMaxCharCountToDisplay() >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.options.GetMaxCharCountToDisplay())
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultConverterTestSuite) TestFormatMapField() {
	mapField := types.MapStatementResultField{
		Type:      types.ARRAY,
		KeyType:   types.VARCHAR,
		ValueType: types.VARCHAR,
		Entries: []types.MapStatementResultFieldEntry{
			{
				Key: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Key1",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Value1",
				},
			},
			{
				Key: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Key2",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Value2",
				},
			},
		},
	}

	testCases := []struct {
		expected string
		options  *types.FormatterOptions
	}{
		{
			expected: "{Key1=Value1, Key2=Value2}",
			options:  nil,
		},
		{
			expected: "...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 0},
		},
		{
			expected: "{Key1=Va...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 11},
		},
		{
			expected: "{Key1=Value1, ...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 17},
		},
		{
			expected: "{Key1=Value1, Key2=Val...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 25},
		},
		{
			expected: "{Key1=Value1, Key2=Value2}",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 26},
		},
	}

	for idx, testCase := range testCases {
		fmt.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := mapField.Format(testCase.options)
		if testCase.options.GetMaxCharCountToDisplay() >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.options.GetMaxCharCountToDisplay())
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultConverterTestSuite) TestFormatRowField() {
	arrayField := types.RowStatementResultField{
		Type:         types.ARRAY,
		ElementTypes: []types.StatementResultFieldType{types.VARCHAR, types.VARCHAR, types.VARCHAR},
		Values: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "Test",
			},
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "Hello",
			},
			types.AtomicStatementResultField{
				Type:  types.VARCHAR,
				Value: "World",
			},
		},
	}

	testCases := []struct {
		expected string
		options  *types.FormatterOptions
	}{
		{
			expected: "(Test, Hello, World)",
			options:  nil,
		},
		{
			expected: "...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 0},
		},
		{
			expected: "(...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 4},
		},
		{
			expected: "(Test, ...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 10},
		},
		{
			expected: "(Test, Hello, Wo...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 19},
		},
		{
			expected: "(Test, Hello, World)",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 20},
		},
	}

	for idx, testCase := range testCases {
		fmt.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := arrayField.Format(testCase.options)
		if testCase.options.GetMaxCharCountToDisplay() >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.options.GetMaxCharCountToDisplay())
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultConverterTestSuite) TestFormatNestedField() {
	mapField := types.MapStatementResultField{
		Type:      types.ARRAY,
		KeyType:   types.VARCHAR,
		ValueType: types.VARCHAR,
		Entries: []types.MapStatementResultFieldEntry{
			{
				Key: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Key1",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Value1",
				},
			},
			{
				Key: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Key2",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.VARCHAR,
					Value: "Value2",
				},
			},
		},
	}

	field := types.ArrayStatementResultField{
		Type:        types.ARRAY,
		ElementType: types.MAP,
		Values: []types.StatementResultField{
			mapField,
			mapField,
		},
	}

	testCases := []struct {
		expected string
		options  *types.FormatterOptions
	}{
		{
			expected: "[{Key1=Value1, Key2=Value2}, {Key1=Value1, Key2=Value2}]",
			options:  nil,
		},
		{
			expected: "...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 0},
		},
		{
			expected: "[...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 4},
		},
		{
			expected: "[{Key1=Value1,...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 17},
		},
		{
			expected: "[{Key1=Value1, Key2=Value2}, {Key1=Value1, Key2=Valu...",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 55},
		},
		{
			expected: "[{Key1=Value1, Key2=Value2}, {Key1=Value1, Key2=Value2}]",
			options:  &types.FormatterOptions{MaxCharCountToDisplay: 56},
		},
	}

	for idx, testCase := range testCases {
		fmt.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := field.Format(testCase.options)
		if testCase.options.GetMaxCharCountToDisplay() >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.options.GetMaxCharCountToDisplay())
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}
