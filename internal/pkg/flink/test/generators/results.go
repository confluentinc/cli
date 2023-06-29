package generators

import (
	"fmt"
	"strconv"
	"time"

	"pgregory.net/rapid"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func GetResultItemGeneratorForType(dataType flinkgatewayv1alpha1.DataType) *rapid.Generator[any] {
	fieldType := types.NewResultFieldType(dataType)
	switch fieldType {
	case types.ARRAY:
		elementType := dataType.GetElementType()
		return ArrayResultItem(elementType)
	case types.MULTISET:
		keyType := dataType.GetElementType()
		valueType := flinkgatewayv1alpha1.DataType{
			Nullable: false,
			Type:     "INTEGER",
		}
		return MapResultItem(keyType, valueType)
	case types.MAP:
		keyType := dataType.GetKeyType()
		valueType := dataType.GetValueType()
		return MapResultItem(keyType, valueType)
	case types.ROW:
		elementTypes := dataType.GetFields()
		return RowResultItem(elementTypes)
	case types.NULL:
		return rapid.SampledFrom([]any{nil})
	default:
		return AtomicResultItem(fieldType)
	}
}

var atomicGenerators = map[types.StatementResultFieldType]*rapid.Generator[string]{
	types.CHAR:                           rapid.SampledFrom([]string{"Jay", "Yannick", "Gustavo", "Jim", "A string", "Another string", "And another string", "lorem ipsum"}),
	types.VARCHAR:                        rapid.SampledFrom([]string{"Jay", "Yannick", "Gustavo", "Jim", "A string", "Another string", "And another string", "lorem ipsum"}),
	types.BOOLEAN:                        rapid.SampledFrom([]string{"TRUE", "FALSE"}),
	types.BINARY:                         rapid.StringMatching("x'[a-fA-F0-9]+'"),
	types.VARBINARY:                      rapid.StringMatching("x'[a-fA-F0-9]+'"),
	types.DECIMAL:                        rapid.StringMatching("\\d{1,5}\\.\\d{1,3}"),
	types.TINYINT:                        rapid.Custom(func(t *rapid.T) string { return fmt.Sprintf("%d", int(rapid.Int8().Draw(t, "an int"))) }),
	types.SMALLINT:                       rapid.Custom(func(t *rapid.T) string { return fmt.Sprintf("%d", int(rapid.Int16().Draw(t, "an int"))) }),
	types.INTEGER:                        rapid.Custom(func(t *rapid.T) string { return fmt.Sprintf("%d", rapid.Int().Draw(t, "an int")) }),
	types.BIGINT:                         rapid.Custom(func(t *rapid.T) string { return fmt.Sprintf("%d", rapid.Int64().Draw(t, "an int")) }),
	types.FLOAT:                          rapid.StringMatching("\\d{1,5}\\.\\d{7}E7"),
	types.DOUBLE:                         rapid.StringMatching("\\d{1,5}\\.\\d{16}E7"),
	types.DATE:                           Timestamp("2006-01-02"),
	types.TIME_WITHOUT_TIME_ZONE:         Timestamp("15:04:05.000000"),
	types.TIMESTAMP_WITHOUT_TIME_ZONE:    Timestamp("2006-01-02 15:04:05.000000"),
	types.TIMESTAMP_WITH_TIME_ZONE:       Timestamp("2006-01-02 15:04:05.000000"),
	types.TIMESTAMP_WITH_LOCAL_TIME_ZONE: Timestamp("2006-01-02 15:04:05.000000"),
	types.INTERVAL_YEAR_MONTH:            rapid.Custom(func(t *rapid.T) string { return "+" + Timestamp("2006-01").Draw(t, "a timestamp") }),
	types.INTERVAL_DAY_TIME: rapid.Custom(func(t *rapid.T) string {
		return "+" + rapid.Custom(func(t *rapid.T) string { return fmt.Sprintf("%d", rapid.IntRange(0, 365).Draw(t, "a day")) }).Draw(t, "a day") + Timestamp("15:04:05.000000").Draw(t, "a timestamp")
	}),
}

func Timestamp(formatString string) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		seconds := rapid.IntRange(0, 31536000).Draw(t, "Timestamp")
		createdAt := time.Now().Add(time.Duration(seconds) * time.Second)
		return createdAt.Format(formatString)
	})
}

// AtomicResultItem generates a random atomic field
func AtomicResultItem(fieldType types.StatementResultFieldType) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		atomicGenerator := atomicGenerators[fieldType]
		return atomicGenerator.Draw(t, "an atomic value")
	})
}

// ArrayResultItem generates a random ARRAY field
func ArrayResultItem(elementDataType flinkgatewayv1alpha1.DataType) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var arrayItems []any
		arraySize := rapid.IntRange(1, 3).Draw(t, "array size")
		elementGenerator := GetResultItemGeneratorForType(elementDataType)
		for i := 0; i < arraySize; i++ {
			arrayItems = append(arrayItems, elementGenerator.Draw(t, "an array item"))
		}
		return arrayItems
	})
}

// MapResultItem generates a random MAP field
func MapResultItem(keyType, valueType flinkgatewayv1alpha1.DataType) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var mapItems []any
		arraySize := rapid.IntRange(1, 3).Draw(t, "map size")
		keyGenerator := GetResultItemGeneratorForType(keyType)
		valueGenerator := GetResultItemGeneratorForType(valueType)
		for i := 0; i < arraySize; i++ {
			var keyValuePair []any
			keyValuePair = append(keyValuePair, keyGenerator.Draw(t, "key"), valueGenerator.Draw(t, "value"))
			mapItems = append(mapItems, keyValuePair)
		}
		return mapItems
	})
}

// RowResultItem generates a random ROW field
func RowResultItem(fieldTypes []flinkgatewayv1alpha1.RowFieldType) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var arrayItems []any
		for i := range fieldTypes {
			generator := GetResultItemGeneratorForType(fieldTypes[i].GetFieldType())
			arrayItems = append(arrayItems, generator.Draw(t, "an array item"))
		}
		return arrayItems
	})
}

// MockResultRow creates a row with random fields adhering to the provided column schema
func MockResultRow(columnDetails []flinkgatewayv1alpha1.ColumnDetails) *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		var items []any
		for _, column := range columnDetails {
			items = append(items, GetResultItemGeneratorForType(column.GetType()).Draw(t, "a field"))
		}
		return map[string]any{
			"op":  rapid.Float64Range(0, 3).Draw(t, "an operation"),
			"row": items,
		}
	})
}

var NonAtomicResultFieldTypes = []types.StatementResultFieldType{
	types.ARRAY,
	types.MULTISET,
	types.MAP,
	types.ROW,
}

var AtomicResultFieldTypes = []types.StatementResultFieldType{
	types.CHAR,
	types.VARCHAR,
	types.BOOLEAN,
	types.BINARY,
	types.VARBINARY,
	types.DECIMAL,
	types.TINYINT,
	types.SMALLINT,
	types.INTEGER,
	types.BIGINT,
	types.FLOAT,
	types.DOUBLE,
	types.DATE,
	types.TIME_WITHOUT_TIME_ZONE,
	types.TIMESTAMP_WITHOUT_TIME_ZONE,
	types.TIMESTAMP_WITH_TIME_ZONE,
	types.TIMESTAMP_WITH_LOCAL_TIME_ZONE,
	types.INTERVAL_YEAR_MONTH,
	types.INTERVAL_DAY_TIME,
	types.NULL,
}

func getDataTypeGeneratorForType(fieldType types.StatementResultFieldType, maxNestingDepth int) *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	if maxNestingDepth <= 0 {
		return AtomicDataType()
	}
	switch fieldType {
	case types.ARRAY:
		return ArrayDataType(maxNestingDepth - 1)
	case types.MULTISET:
		return MultisetDataType(maxNestingDepth - 1)
	case types.MAP:
		return MapDataType(maxNestingDepth - 1)
	case types.ROW:
		return RowDataType(maxNestingDepth - 1)
	default:
		return AtomicDataType()
	}
}

// AtomicDataType generates a random atomic data type
func AtomicDataType() *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	return rapid.Custom(func(t *rapid.T) flinkgatewayv1alpha1.DataType {
		resultFieldType := rapid.SampledFrom(AtomicResultFieldTypes).Draw(t, "atomic result field type")
		dataTypeJson := fmt.Sprintf(`{"type": "%s"}`, string(resultFieldType))
		dataType := flinkgatewayv1alpha1.NewNullableDataType(nil)
		if err := dataType.UnmarshalJSON([]byte(dataTypeJson)); err != nil {
			return flinkgatewayv1alpha1.DataType{}
		}
		return *dataType.Get()
	})
}

// ArrayDataType generates a random array data type
func ArrayDataType(maxNestingDepth int) *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	return rapid.Custom(func(t *rapid.T) flinkgatewayv1alpha1.DataType {
		resultFieldType := GenResultFieldType().Draw(t, "result field type")
		elementType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "element type")
		return flinkgatewayv1alpha1.DataType{
			Nullable:    false,
			Type:        "ARRAY",
			ElementType: &elementType,
		}
	})
}

// MapDataType generates a random map data type
func MapDataType(maxNestingDepth int) *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	return rapid.Custom(func(t *rapid.T) flinkgatewayv1alpha1.DataType {
		resultFieldKeyType := GenResultFieldType().Draw(t, "result field type")
		resultFieldValueType := GenResultFieldType().Draw(t, "result field type")
		keyType := getDataTypeGeneratorForType(resultFieldKeyType, maxNestingDepth).Draw(t, "element type")
		valueType := getDataTypeGeneratorForType(resultFieldValueType, maxNestingDepth).Draw(t, "element type")
		return flinkgatewayv1alpha1.DataType{
			Nullable:  false,
			Type:      "MAP",
			KeyType:   &keyType,
			ValueType: &valueType,
		}
	})
}

// MultisetDataType generates a random map data type
func MultisetDataType(maxNestingDepth int) *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	return rapid.Custom(func(t *rapid.T) flinkgatewayv1alpha1.DataType {
		resultFieldType := GenResultFieldType().Draw(t, "result field type")
		elementType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "element type")
		return flinkgatewayv1alpha1.DataType{
			Nullable:    false,
			Type:        "MULTISET",
			ElementType: &elementType,
		}
	})
}

// RowDataType generates a random row data type
func RowDataType(maxNestingDepth int) *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	return rapid.Custom(func(t *rapid.T) flinkgatewayv1alpha1.DataType {
		var fieldTypes []flinkgatewayv1alpha1.RowFieldType
		rowSize := rapid.IntRange(1, 3).Draw(t, "array size")
		for i := 0; i < rowSize; i++ {
			resultFieldType := GenResultFieldType().Draw(t, "result field type")
			elementType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "element type")
			fieldTypes = append(fieldTypes, flinkgatewayv1alpha1.RowFieldType{
				Name:      strconv.Itoa(i),
				FieldType: elementType,
			})
		}
		return flinkgatewayv1alpha1.DataType{
			Nullable: false,
			Type:     "ROW",
			Fields:   &fieldTypes,
		}
	})
}

func GenResultFieldType() *rapid.Generator[types.StatementResultFieldType] {
	return rapid.Custom(func(t *rapid.T) types.StatementResultFieldType {
		// this should about even the chances for an atomic vs. non-atomic field
		shouldGenAtomic := rapid.Bool().Draw(t, "bool")
		if shouldGenAtomic {
			return rapid.SampledFrom(AtomicResultFieldTypes).Draw(t, "atomic result field type")
		}
		return rapid.SampledFrom(NonAtomicResultFieldTypes).Draw(t, "result field type")
	})
}

func DataType(maxNestingDepth int) *rapid.Generator[flinkgatewayv1alpha1.DataType] {
	return rapid.Custom(func(t *rapid.T) flinkgatewayv1alpha1.DataType {
		resultFieldType := GenResultFieldType().Draw(t, "result field type")
		return getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "data type")
	})
}

func MockResultColumns(numColumns, maxNestingDepth int) *rapid.Generator[[]flinkgatewayv1alpha1.ColumnDetails] {
	return rapid.Custom(func(t *rapid.T) []flinkgatewayv1alpha1.ColumnDetails {
		var columnDetails []flinkgatewayv1alpha1.ColumnDetails
		for i := 0; i < numColumns; i++ {
			dataType := DataType(maxNestingDepth).Draw(t, "column type")
			columnDetails = append(columnDetails, flinkgatewayv1alpha1.ColumnDetails{
				Name: string(types.NewResultFieldType(dataType)),
				Type: dataType,
			})
		}
		return columnDetails
	})
}

func MockResults(maxNumColumns, maxNestingDepth int) *rapid.Generator[types.MockStatementResult] {
	return rapid.Custom(func(t *rapid.T) types.MockStatementResult {
		if maxNumColumns <= 0 {
			maxNumColumns = 10
		}
		if maxNestingDepth < 0 {
			maxNestingDepth = 10
		}
		maxNumColumns = rapid.IntRange(1, maxNumColumns).Draw(t, "column number")
		columnDetails := MockResultColumns(maxNumColumns, maxNestingDepth).Draw(t, "column details")
		resultData := rapid.SliceOfN(MockResultRow(columnDetails), 20, 50).Draw(t, "result data")

		return types.MockStatementResult{
			ResultSchema: flinkgatewayv1alpha1.SqlV1alpha1ResultSchema{Columns: &columnDetails},
			StatementResults: flinkgatewayv1alpha1.SqlV1alpha1StatementResult{
				Results: &flinkgatewayv1alpha1.SqlV1alpha1StatementResultResults{Data: &resultData},
			},
		}
	})
}
