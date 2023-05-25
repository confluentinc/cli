package results

import (
	"fmt"
	"github.com/samber/lo"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type ResultFormatterTestSuite struct {
	suite.Suite
}

func TestResultFormatterTestSuite(t *testing.T) {
	suite.Run(t, new(ResultFormatterTestSuite))
}

func (s *ResultFormatterTestSuite) TestGetTruncatedColumnWidthsShouldMaxOutAvailableSpace() {
	rapid.Check(s.T(), func(t *rapid.T) {
		columnWidths := rapid.SliceOfN(rapid.IntRange(0, 40), 1, 10).Draw(t, "column widths")
		maxCharacters := rapid.IntRange(40, 150).Draw(t, "max characters")
		truncatedColumnWidths := GetTruncatedColumnWidths(columnWidths, maxCharacters)

		if maxCharacters >= lo.Sum(columnWidths) {
			// no truncation occurred -> columns should not have changed
			require.Equal(t, columnWidths, truncatedColumnWidths)
		} else {
			// truncation occurred -> check if available space is maxed out
			require.Equal(t, maxCharacters, lo.Sum(truncatedColumnWidths))
		}
	})
}

func (s *ResultFormatterTestSuite) TestGetTruncatedColumnWidthsShouldNotAssignColumnsMoreThanTheyNeed() {
	rapid.Check(s.T(), func(t *rapid.T) {
		columnWidths := rapid.SliceOfN(rapid.IntRange(0, 40), 1, 10).Draw(t, "column widths")
		maxCharacters := rapid.IntRange(40, 150).Draw(t, "max characters")
		truncatedColumnWidths := GetTruncatedColumnWidths(columnWidths, maxCharacters)

		for colIdx, truncatedColumnWidth := range truncatedColumnWidths {
			require.LessOrEqual(t, truncatedColumnWidth, columnWidths[colIdx])
		}
	})
}

func (s *ResultFormatterTestSuite) TestGetTruncatedColumnWidthsDistributesLeftoverSpaceGreedily() {
	testCases := []struct {
		columnWidths                  []int
		maxCharacters                 int
		expectedTruncatedColumnWidths []int
	}{
		{columnWidths: []int{20, 20, 20}, maxCharacters: 30, expectedTruncatedColumnWidths: []int{10, 10, 10}},
		{columnWidths: []int{10, 20, 20}, maxCharacters: 30, expectedTruncatedColumnWidths: []int{10, 10, 10}},
		{columnWidths: []int{8, 20, 20}, maxCharacters: 30, expectedTruncatedColumnWidths: []int{8, 12, 10}},
		{columnWidths: []int{1, 18, 20}, maxCharacters: 30, expectedTruncatedColumnWidths: []int{1, 18, 11}},
	}

	for idx, tc := range testCases {
		output.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		require.Equal(s.T(), tc.expectedTruncatedColumnWidths, GetTruncatedColumnWidths(tc.columnWidths, tc.maxCharacters))
	}
}

func (s *ResultFormatterTestSuite) TestFormatAtomicField() {
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

		require.Equal(s.T(), val, convertedField.ToString())
		maxDisplayableCharCount := rapid.IntRange(-3, 20).Draw(t, "max displayable chars")
		if len(val) > maxDisplayableCharCount {
			if maxDisplayableCharCount <= 3 {
				require.Equal(s.T(), "...", TruncateString(convertedField.ToString(), maxDisplayableCharCount))
			} else {
				require.Equal(s.T(), val[:maxDisplayableCharCount-3]+"...", TruncateString(convertedField.ToString(), maxDisplayableCharCount))
			}
		} else {
			require.Equal(s.T(), val, TruncateString(convertedField.ToString(), maxDisplayableCharCount))
		}
	})
}

func (s *ResultFormatterTestSuite) TestFormatArrayField() {
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
		expected              string
		maxCharCountToDisplay int
	}{
		{
			expected:              "...",
			maxCharCountToDisplay: 0,
		},
		{
			expected:              "[...",
			maxCharCountToDisplay: 4,
		},
		{
			expected:              "[Test, ...",
			maxCharCountToDisplay: 10,
		},
		{
			expected:              "[Test, Hello, Wo...",
			maxCharCountToDisplay: 19,
		},
		{
			expected:              "[Test, Hello, World]",
			maxCharCountToDisplay: 20,
		},
	}

	for idx, testCase := range testCases {
		output.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := TruncateString(arrayField.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestFormatMapField() {
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
		expected              string
		maxCharCountToDisplay int
	}{
		{
			expected:              "...",
			maxCharCountToDisplay: 0,
		},
		{
			expected:              "{Key1=Va...",
			maxCharCountToDisplay: 11,
		},
		{
			expected:              "{Key1=Value1, ...",
			maxCharCountToDisplay: 17,
		},
		{
			expected:              "{Key1=Value1, Key2=Val...",
			maxCharCountToDisplay: 25,
		},
		{
			expected:              "{Key1=Value1, Key2=Value2}",
			maxCharCountToDisplay: 26,
		},
	}

	for idx, testCase := range testCases {
		output.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := TruncateString(mapField.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestFormatRowField() {
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
		expected              string
		maxCharCountToDisplay int
	}{
		{
			expected:              "...",
			maxCharCountToDisplay: 0,
		},
		{
			expected:              "(...",
			maxCharCountToDisplay: 4,
		},
		{
			expected:              "(Test, ...",
			maxCharCountToDisplay: 10,
		},
		{
			expected:              "(Test, Hello, Wo...",
			maxCharCountToDisplay: 19,
		},
		{
			expected:              "(Test, Hello, World)",
			maxCharCountToDisplay: 20,
		},
	}

	for idx, testCase := range testCases {
		output.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := TruncateString(arrayField.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestFormatNestedField() {
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
		expected              string
		maxCharCountToDisplay int
	}{
		{
			expected:              "...",
			maxCharCountToDisplay: 0,
		},
		{
			expected:              "[...",
			maxCharCountToDisplay: 4,
		},
		{
			expected:              "[{Key1=Value1,...",
			maxCharCountToDisplay: 17,
		},
		{
			expected:              "[{Key1=Value1, Key2=Value2}, {Key1=Value1, Key2=Valu...",
			maxCharCountToDisplay: 55,
		},
		{
			expected:              "[{Key1=Value1, Key2=Value2}, {Key1=Value1, Key2=Value2}]",
			maxCharCountToDisplay: 56,
		},
	}

	for idx, testCase := range testCases {
		output.Println(fmt.Sprintf("Evaluating test case #%v", idx))
		formattedField := TruncateString(field.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}
