package generators

import (
	"fmt"
	"strconv"

	"pgregory.net/rapid"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

func GetResultItemGeneratorForTypeOnPrem(dataType cmfsdk.DataType) *rapid.Generator[any] {
	fieldType := types.NewResultFieldType(dataType.GetType())
	switch fieldType {
	case types.Array:
		elementType := dataType.GetElementType()
		return ArrayResultItemOnPrem(elementType)
	case types.Multiset:
		keyType := dataType.GetElementType()
		valueType := cmfsdk.DataType{
			Nullable: false,
			Type:     "INTEGER",
		}
		return MapResultItemOnPrem(keyType, valueType)
	case types.Map:
		keyType := dataType.GetKeyType()
		valueType := dataType.GetValueType()
		return MapResultItemOnPrem(keyType, valueType)
	case types.Row:
		elementTypes := dataType.GetFields()
		return RowResultItemOnPrem(elementTypes)
	case types.Null:
		return rapid.SampledFrom([]any{nil})
	default:
		return AtomicResultItem(fieldType)
	}
}

// ArrayResultItemOnPrem generates a random ARRAY field
func ArrayResultItemOnPrem(elementDataType cmfsdk.DataType) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var arrayItems []any
		arraySize := rapid.IntRange(1, 3).Draw(t, "array size")
		elementGenerator := GetResultItemGeneratorForTypeOnPrem(elementDataType)
		for i := 0; i < arraySize; i++ {
			arrayItems = append(arrayItems, elementGenerator.Draw(t, "an array item"))
		}
		return arrayItems
	})
}

// MapResultItemOnPrem generates a random MAP field
func MapResultItemOnPrem(keyType, valueType cmfsdk.DataType) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var mapItems []any
		arraySize := rapid.IntRange(1, 3).Draw(t, "map size")
		keyGenerator := GetResultItemGeneratorForTypeOnPrem(keyType)
		valueGenerator := GetResultItemGeneratorForTypeOnPrem(valueType)
		for i := 0; i < arraySize; i++ {
			var keyValuePair []any
			keyValuePair = append(keyValuePair, keyGenerator.Draw(t, "key"), valueGenerator.Draw(t, "value"))
			mapItems = append(mapItems, keyValuePair)
		}
		return mapItems
	})
}

// RowResultItemOnPrem generates a random ROW field
func RowResultItemOnPrem(fieldTypes []cmfsdk.DataTypeField) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var arrayItems []any
		for i := range fieldTypes {
			generator := GetResultItemGeneratorForTypeOnPrem(fieldTypes[i].GetFieldType())
			arrayItems = append(arrayItems, generator.Draw(t, "an array item"))
		}
		return arrayItems
	})
}

// MockResultRow creates a row with random fields adhering to the provided column schema
func MockResultRowOnPrem(columnDetails []cmfsdk.ResultSchemaColumn) *rapid.Generator[map[string]interface{}] {
	return rapid.Custom(func(t *rapid.T) map[string]interface{} {
		var items []any
		for _, column := range columnDetails {
			items = append(items, GetResultItemGeneratorForTypeOnPrem(column.GetType()).Draw(t, "a field"))
		}
		return map[string]any{
			"op":  float64(rapid.IntRange(0, 3).Draw(t, "an operation")),
			"row": items,
		}
	})
}

func getDataTypeGeneratorForTypeOnPrem(fieldType types.StatementResultFieldType, maxNestingDepth int) *rapid.Generator[cmfsdk.DataType] {
	if maxNestingDepth <= 0 {
		return AtomicDataTypeOnPrem()
	}
	switch fieldType {
	case types.Array:
		return ArrayDataTypeOnPrem(maxNestingDepth - 1)
	case types.Multiset:
		return MultisetDataTypeOnPrem(maxNestingDepth - 1)
	case types.Map:
		return MapDataTypeOnPrem(maxNestingDepth - 1)
	case types.Row:
		return RowDataTypeOnPrem(maxNestingDepth - 1)
	default:
		return AtomicDataTypeOnPrem()
	}
}

// AtomicDataTypeOnPrem generates a random atomic data type
func AtomicDataTypeOnPrem() *rapid.Generator[cmfsdk.DataType] {
	return rapid.Custom(func(t *rapid.T) cmfsdk.DataType {
		resultFieldType := rapid.SampledFrom(AtomicResultFieldTypes).Draw(t, "atomic result field type")
		dataTypeJson := fmt.Sprintf(`{"type": "%s"}`, string(resultFieldType))
		dataType := cmfsdk.NewNullableDataType(nil)
		if err := dataType.UnmarshalJSON([]byte(dataTypeJson)); err != nil {
			return cmfsdk.DataType{}
		}
		return *dataType.Get()
	})
}

// ArrayDataTypeOnPrem generates a random array data type
func ArrayDataTypeOnPrem(maxNestingDepth int) *rapid.Generator[cmfsdk.DataType] {
	return rapid.Custom(func(t *rapid.T) cmfsdk.DataType {
		resultFieldType := GenResultFieldType().Draw(t, resultFieldLabel)
		elementType := getDataTypeGeneratorForTypeOnPrem(resultFieldType, maxNestingDepth).Draw(t, elementLabel)
		return cmfsdk.DataType{
			Nullable:    false,
			Type:        "ARRAY",
			ElementType: &elementType,
		}
	})
}

// MapDataTypeOnPrem generates a random map data type
func MapDataTypeOnPrem(maxNestingDepth int) *rapid.Generator[cmfsdk.DataType] {
	return rapid.Custom(func(t *rapid.T) cmfsdk.DataType {
		resultFieldKeyType := GenResultFieldType().Draw(t, resultFieldLabel)
		resultFieldValueType := GenResultFieldType().Draw(t, resultFieldLabel)
		keyType := getDataTypeGeneratorForTypeOnPrem(resultFieldKeyType, maxNestingDepth).Draw(t, elementLabel)
		valueType := getDataTypeGeneratorForTypeOnPrem(resultFieldValueType, maxNestingDepth).Draw(t, elementLabel)
		return cmfsdk.DataType{
			Nullable:  false,
			Type:      "MAP",
			KeyType:   &keyType,
			ValueType: &valueType,
		}
	})
}

// MultisetDataTypeOnPrem generates a random map data type
func MultisetDataTypeOnPrem(maxNestingDepth int) *rapid.Generator[cmfsdk.DataType] {
	return rapid.Custom(func(t *rapid.T) cmfsdk.DataType {
		resultFieldType := GenResultFieldType().Draw(t, resultFieldLabel)
		elementType := getDataTypeGeneratorForTypeOnPrem(resultFieldType, maxNestingDepth).Draw(t, elementLabel)
		return cmfsdk.DataType{
			Nullable:    false,
			Type:        "MULTISET",
			ElementType: &elementType,
		}
	})
}

// RowDataTypeOnPrem generates a random row data type
func RowDataTypeOnPrem(maxNestingDepth int) *rapid.Generator[cmfsdk.DataType] {
	return rapid.Custom(func(t *rapid.T) cmfsdk.DataType {
		var fieldTypes []cmfsdk.DataTypeField
		rowSize := rapid.IntRange(1, 3).Draw(t, "array size")
		for i := 0; i < rowSize; i++ {
			resultFieldType := GenResultFieldType().Draw(t, resultFieldLabel)
			elementType := getDataTypeGeneratorForTypeOnPrem(resultFieldType, maxNestingDepth).Draw(t, elementLabel)
			fieldTypes = append(fieldTypes, cmfsdk.DataTypeField{
				Name:      strconv.Itoa(i),
				FieldType: elementType,
			})
		}
		return cmfsdk.DataType{
			Nullable: false,
			Type:     "ROW",
			Fields:   &fieldTypes,
		}
	})
}

func DataTypeOnPrem(maxNestingDepth int) *rapid.Generator[cmfsdk.DataType] {
	return rapid.Custom(func(t *rapid.T) cmfsdk.DataType {
		resultFieldType := GenResultFieldType().Draw(t, resultFieldLabel)
		return getDataTypeGeneratorForTypeOnPrem(resultFieldType, maxNestingDepth).Draw(t, "data type")
	})
}

func MockResultColumnsOnPrem(numColumns, maxNestingDepth int) *rapid.Generator[[]cmfsdk.ResultSchemaColumn] {
	return rapid.Custom(func(t *rapid.T) []cmfsdk.ResultSchemaColumn {
		var columnDetails []cmfsdk.ResultSchemaColumn
		for i := 0; i < numColumns; i++ {
			dataType := DataTypeOnPrem(maxNestingDepth).Draw(t, "column type")
			columnDetails = append(columnDetails, cmfsdk.ResultSchemaColumn{
				Name: string(types.NewResultFieldType(dataType.GetType())),
				Type: dataType,
			})
		}
		return columnDetails
	})
}

func MockResultsOnPrem(maxNumColumns, maxNestingDepth int) *rapid.Generator[types.MockStatementResultOnPrem] {
	return rapid.Custom(func(t *rapid.T) types.MockStatementResultOnPrem {
		if maxNumColumns <= 0 {
			maxNumColumns = 10
		}
		if maxNestingDepth < 0 {
			maxNestingDepth = 10
		}
		maxNumColumns = rapid.IntRange(1, maxNumColumns).Draw(t, "column number")
		columnDetails := MockResultColumnsOnPrem(maxNumColumns, maxNestingDepth).Draw(t, "column details")
		resultData := rapid.SliceOfN(MockResultRowOnPrem(columnDetails), 20, 50).Draw(t, "result data")

		return types.MockStatementResultOnPrem{
			ResultSchema: cmfsdk.ResultSchema{Columns: columnDetails},
			StatementResults: cmfsdk.StatementResult{
				Results: cmfsdk.StatementResults{Data: &resultData},
			},
		}
	})
}
