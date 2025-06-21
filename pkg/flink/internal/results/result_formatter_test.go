package results

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/test/generators"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
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
		fmt.Printf("Evaluating test case #%d\n", idx)
		require.Equal(s.T(), tc.expectedTruncatedColumnWidths, GetTruncatedColumnWidths(tc.columnWidths, tc.maxCharacters))
	}
}

func (s *ResultFormatterTestSuite) TestFormatAtomicField() {
	rapid.Check(s.T(), func(t *rapid.T) {
		atomicDataType := generators.AtomicDataType().Draw(t, "atomic data type")
		atomicField := generators.GetResultItemGeneratorForType(atomicDataType).Draw(t, "atomic result field")
		convertedField := convertToInternalField(atomicField, flinkgatewayv1.ColumnDetails{
			Name: "Test_Column",
			Type: atomicDataType,
		})

		val := "NULL"
		if types.NewResultFieldType(atomicDataType.GetType()) != types.Null {
			val, _ = atomicField.(string)
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
		Type:        types.Array,
		ElementType: types.Varchar,
		Values: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: "Test",
			},
			types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: "Hello",
			},
			types.AtomicStatementResultField{
				Type:  types.Varchar,
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
		fmt.Printf("Evaluating test case #%d\n", idx)
		formattedField := TruncateString(arrayField.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestFormatMapField() {
	mapField := types.MapStatementResultField{
		Type:      types.Array,
		KeyType:   types.Varchar,
		ValueType: types.Varchar,
		Entries: []types.MapStatementResultFieldEntry{
			{
				Key: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Key1",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Value1",
				},
			},
			{
				Key: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Key2",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.Varchar,
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
		fmt.Printf("Evaluating test case #%d\n", idx)
		formattedField := TruncateString(mapField.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestFormatRowField() {
	arrayField := types.RowStatementResultField{
		Type:         types.Array,
		ElementTypes: []types.StatementResultFieldType{types.Varchar, types.Varchar, types.Varchar},
		Values: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: "Test",
			},
			types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: "Hello",
			},
			types.AtomicStatementResultField{
				Type:  types.Varchar,
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
		fmt.Printf("Evaluating test case #%d\n", idx)
		formattedField := TruncateString(arrayField.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestFormatNestedField() {
	mapField := types.MapStatementResultField{
		Type:      types.Array,
		KeyType:   types.Varchar,
		ValueType: types.Varchar,
		Entries: []types.MapStatementResultFieldEntry{
			{
				Key: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Key1",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Value1",
				},
			},
			{
				Key: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Key2",
				},
				Value: types.AtomicStatementResultField{
					Type:  types.Varchar,
					Value: "Value2",
				},
			},
		},
	}

	field := types.ArrayStatementResultField{
		Type:        types.Array,
		ElementType: types.Map,
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
		fmt.Printf("Evaluating test case #%d\n", idx)
		formattedField := TruncateString(field.ToString(), testCase.maxCharCountToDisplay)
		if testCase.maxCharCountToDisplay >= 3 {
			require.True(s.T(), len(formattedField) <= testCase.maxCharCountToDisplay)
		}
		require.Equal(s.T(), testCase.expected, formattedField)
	}
}

func (s *ResultFormatterTestSuite) TestTruncateMultiLineStringShouldNotTruncate() {
	testCases := []struct {
		input                 string
		expected              string
		maxCharCountToDisplay int
	}{
		{
			input:                 "SELECT * FROM \n",
			expected:              "SELECT * FROM \n",
			maxCharCountToDisplay: 15,
		},
		{
			input:                 "SELECT * FROM \n table",
			expected:              "SELECT * FROM \n table",
			maxCharCountToDisplay: 15,
		},
		{
			input:                 "SELECT * FROM \n table",
			expected:              "SELECT ...\n table",
			maxCharCountToDisplay: 10,
		},
		{
			input:                 "SELECT * FROM \n table-with-a-really-long-table-name",
			expected:              "SELECT * FROM \n table-with-a-rea...",
			maxCharCountToDisplay: 20,
		},
	}

	for idx, testCase := range testCases {
		fmt.Printf("Evaluating test case #%d\n", idx)
		require.Equal(s.T(), testCase.expected, TruncateString(testCase.input, testCase.maxCharCountToDisplay))
	}
}

func (s *ResultFormatterTestSuite) TestTruncateMultibyteCharacters() {
	testCases := []struct {
		input                 string
		expected              string
		maxCharCountToDisplay int
	}{
		{
			input:                 "あああ",  // The string width here is 6 since each character is 2 bytes wide.
			expected:              "あ...", // あ... is exactly 5 bytes wide.
			maxCharCountToDisplay: 5,
		},
		{
			input:                 "SELECT `あ` FROM ",
			expected:              "SELECT `あ` FROM ",
			maxCharCountToDisplay: 17,
		},
		{
			input:                 "SELECT `あ` FROM \n table",
			expected:              "SELECT `あ` FROM \n table",
			maxCharCountToDisplay: 17,
		},
		{
			input:                 "SELECT `あ` FROM \n table",
			expected:              "SELECT ...\n table",
			maxCharCountToDisplay: 10,
		},
		{
			input:                 "SELECT `あああああああああああああ`",
			expected:              "SELECT `ああああ...",
			maxCharCountToDisplay: 20,
		},
		{
			input:                 "SELECT `あ` FROM \n `🎄🚀👍😀🎄🚀👍🎄🚀👍😀🎄🚀👍`",
			expected:              "SELECT `あ` FROM \n `🎄🚀👍😀🎄🚀👍...",
			maxCharCountToDisplay: 20,
		},
		// There are non-standard emojis like the ones below that can be represented as 1 or 2 width characters
		{
			input:                 "⚒️⚖️",
			expected:              "⚒️⚖️",
			maxCharCountToDisplay: 2,
		},
	}

	for idx, testCase := range testCases {
		fmt.Printf("Evaluating test case #%d\n", idx)
		require.Equal(s.T(), testCase.expected, TruncateString(testCase.input, testCase.maxCharCountToDisplay))
	}
}
