package generators

import (
	"fmt"
	"strconv"
	"time"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"pgregory.net/rapid"
)

func GetResultItemGeneratorForType(dataType v1.DataType) *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	fieldType := types.NewResultFieldType(dataType)
	switch fieldType {
	case types.ARRAY:
		elementType := dataType.ArrayType.GetArrayElementType()
		return ArrayResultItem(elementType)
	case types.MULTISET:
		keyType := dataType.MultisetType.GetMultisetElementType()
		valueType := v1.DataType{IntegerType: &v1.IntegerType{
			Nullable: false,
			Type:     "INTEGER",
		}}
		return MapResultItem(keyType, valueType)
	case types.MAP:
		keyType := dataType.MapType.GetKeyType()
		valueType := dataType.MapType.GetValueType()
		return MapResultItem(keyType, valueType)
	case types.ROW:
		elementTypes := dataType.RowType.GetFields()
		return RowResultItem(elementTypes)
	case types.NULL:
		return rapid.SampledFrom([]v1.SqlV1alpha1ResultItemRowOneOf{{}})
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
		formattedDate := createdAt.Format(formatString)
		return formattedDate
	})
}

// AtomicResultItem generates a random atomic field
func AtomicResultItem(fieldType types.StatementResultFieldType) *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	return rapid.Custom(func(t *rapid.T) v1.SqlV1alpha1ResultItemRowOneOf {
		atomicGenerator := atomicGenerators[fieldType]
		atomicValue := v1.SqlV1alpha1ResultItemString(atomicGenerator.Draw(t, "an atomic value"))
		return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemString: &atomicValue}
	})
}

// ArrayResultItem generates a random ARRAY field
func ArrayResultItem(elementDataType v1.DataType) *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	return rapid.Custom(func(t *rapid.T) v1.SqlV1alpha1ResultItemRowOneOf {
		var arrayItems []v1.SqlV1alpha1ResultItemRowOneOf
		arraySize := rapid.IntRange(1, 3).Draw(t, "array size")
		elementGenerator := GetResultItemGeneratorForType(elementDataType)
		for i := 0; i < arraySize; i++ {
			arrayItems = append(arrayItems, elementGenerator.Draw(t, "an array item"))
		}
		return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: arrayItems}}
	})
}

// MapResultItem generates a random MAP field
func MapResultItem(keyType, valueType v1.DataType) *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	return rapid.Custom(func(t *rapid.T) v1.SqlV1alpha1ResultItemRowOneOf {
		var mapItems []v1.SqlV1alpha1ResultItemRowOneOf
		arraySize := rapid.IntRange(1, 3).Draw(t, "map size")
		keyGenerator := GetResultItemGeneratorForType(keyType)
		valueGenerator := GetResultItemGeneratorForType(valueType)
		for i := 0; i < arraySize; i++ {
			var keyValuePair []v1.SqlV1alpha1ResultItemRowOneOf
			keyValuePair = append(keyValuePair, keyGenerator.Draw(t, "key"), valueGenerator.Draw(t, "value"))
			entry := v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: keyValuePair}}
			mapItems = append(mapItems, entry)
		}
		return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: mapItems}}
	})
}

// RowResultItem generates a random ROW field
func RowResultItem(fieldTypes []v1.RowFieldType) *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	return rapid.Custom(func(t *rapid.T) v1.SqlV1alpha1ResultItemRowOneOf {
		var arrayItems []v1.SqlV1alpha1ResultItemRowOneOf
		for i := range fieldTypes {
			generator := GetResultItemGeneratorForType(fieldTypes[i].GetType())
			arrayItems = append(arrayItems, generator.Draw(t, "an array item"))
		}
		return v1.SqlV1alpha1ResultItemRowOneOf{SqlV1alpha1ResultItemRow: &v1.SqlV1alpha1ResultItemRow{Items: arrayItems}}
	})
}

// MockResultRow creates a row with random fields adhering to the provided column schema
func MockResultRow(columnDetails []v1.ColumnDetails) *rapid.Generator[v1.SqlV1alpha1ResultItem] {
	return rapid.Custom(func(t *rapid.T) v1.SqlV1alpha1ResultItem {
		var items []v1.SqlV1alpha1ResultItemRowOneOf
		for _, column := range columnDetails {
			items = append(items, GetResultItemGeneratorForType(column.GetType()).Draw(t, "a field"))
		}
		op := rapid.Int32Range(0, 3).Draw(t, "an operation")
		return v1.SqlV1alpha1ResultItem{
			Op:  &op,
			Row: v1.SqlV1alpha1ResultItemRow{Items: items},
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

func getDataTypeGeneratorForType(fieldType types.StatementResultFieldType, maxNestingDepth int) *rapid.Generator[v1.DataType] {
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
func AtomicDataType() *rapid.Generator[v1.DataType] {
	return rapid.Custom(func(t *rapid.T) v1.DataType {
		resultFieldType := rapid.SampledFrom(AtomicResultFieldTypes).Draw(t, "atomic result field type")
		dataTypeJson := fmt.Sprintf(`{"type": "%s"}`, string(resultFieldType))
		dataType := v1.NewNullableDataType(nil)
		dataType.UnmarshalJSON([]byte(dataTypeJson))
		return *dataType.Get()
	})
}

// ArrayDataType generates a random array data type
func ArrayDataType(maxNestingDepth int) *rapid.Generator[v1.DataType] {
	return rapid.Custom(func(t *rapid.T) v1.DataType {
		resultFieldType := GenResultFieldType().Draw(t, "result field type")
		elementType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "element type")
		return v1.ArrayTypeAsDataType(&v1.ArrayType{
			Nullable:         false,
			Type:             "ARRAY",
			ArrayElementType: elementType,
		})
	})
}

// MapDataType generates a random map data type
func MapDataType(maxNestingDepth int) *rapid.Generator[v1.DataType] {
	return rapid.Custom(func(t *rapid.T) v1.DataType {
		resultFieldKeyType := GenResultFieldType().Draw(t, "result field type")
		resultFieldValueType := GenResultFieldType().Draw(t, "result field type")
		keyType := getDataTypeGeneratorForType(resultFieldKeyType, maxNestingDepth).Draw(t, "element type")
		valueType := getDataTypeGeneratorForType(resultFieldValueType, maxNestingDepth).Draw(t, "element type")
		return v1.MapTypeAsDataType(&v1.MapType{
			Nullable:  false,
			Type:      "MAP",
			KeyType:   keyType,
			ValueType: valueType,
		})
	})
}

// MultisetDataType generates a random map data type
func MultisetDataType(maxNestingDepth int) *rapid.Generator[v1.DataType] {
	return rapid.Custom(func(t *rapid.T) v1.DataType {
		resultFieldType := GenResultFieldType().Draw(t, "result field type")
		elementType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "element type")
		return v1.MultisetTypeAsDataType(&v1.MultisetType{
			Nullable:            false,
			Type:                "MULTISET",
			MultisetElementType: elementType,
		})
	})
}

// RowDataType generates a random row data type
func RowDataType(maxNestingDepth int) *rapid.Generator[v1.DataType] {
	return rapid.Custom(func(t *rapid.T) v1.DataType {
		var fieldTypes []v1.RowFieldType
		rowSize := rapid.IntRange(1, 3).Draw(t, "array size")
		for i := 0; i < rowSize; i++ {
			resultFieldType := GenResultFieldType().Draw(t, "result field type")
			elementType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "element type")
			fieldTypes = append(fieldTypes, v1.RowFieldType{
				Name: strconv.Itoa(i),
				Type: elementType,
			})
		}
		return v1.RowTypeAsDataType(&v1.RowType{
			Nullable: false,
			Type:     "ROW",
			Fields:   fieldTypes,
		})
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

func DataType(maxNestingDepth int) *rapid.Generator[v1.DataType] {
	return rapid.Custom(func(t *rapid.T) v1.DataType {
		resultFieldType := GenResultFieldType().Draw(t, "result field type")
		dataType := getDataTypeGeneratorForType(resultFieldType, maxNestingDepth).Draw(t, "data type")
		return dataType
	})
}

func MockResultColumns(numColumns, maxNestingDepth int) *rapid.Generator[[]v1.ColumnDetails] {
	return rapid.Custom(func(t *rapid.T) []v1.ColumnDetails {
		var columnDetails []v1.ColumnDetails
		for i := 0; i < numColumns; i++ {
			dataType := DataType(maxNestingDepth).Draw(t, "column type")
			columnDetails = append(columnDetails, v1.ColumnDetails{
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
			ResultSchema: v1.SqlV1alpha1ResultSchema{Columns: &columnDetails},
			StatementResults: v1.SqlV1alpha1StatementResult{
				Results: &v1.SqlV1alpha1StatementResultResults{Data: &resultData},
			},
		}
	})
}

func MockCount(count int) types.MockStatementResult {
	var columnDetails []v1.ColumnDetails
	dataType := v1.IntegerTypeAsDataType(&v1.IntegerType{
		Nullable: false,
		Type:     "INTEGER",
	})
	columnDetails = append(columnDetails, v1.ColumnDetails{
		Name: "Count",
		Type: dataType,
	})

	var resultData []v1.SqlV1alpha1ResultItem
	if count == 0 {
		str := v1.SqlV1alpha1ResultItemString(fmt.Sprintf("%v", count))
		op := int32(0)
		item := v1.SqlV1alpha1ResultItem{
			Op: &op,
			Row: v1.SqlV1alpha1ResultItemRow{Items: []v1.SqlV1alpha1ResultItemRowOneOf{
				{SqlV1alpha1ResultItemString: &str},
			}},
		}
		resultData = append(resultData, item)
	} else {
		updateBefore := int32(1)
		valBefore := v1.SqlV1alpha1ResultItemString(fmt.Sprintf("%v", count-1))
		resultData = append(resultData, v1.SqlV1alpha1ResultItem{
			Op: &updateBefore,
			Row: v1.SqlV1alpha1ResultItemRow{Items: []v1.SqlV1alpha1ResultItemRowOneOf{
				{SqlV1alpha1ResultItemString: &valBefore},
			}},
		})

		updateAfter := int32(2)
		valAfter := v1.SqlV1alpha1ResultItemString(fmt.Sprintf("%v", count))
		resultData = append(resultData, v1.SqlV1alpha1ResultItem{
			Op: &updateAfter,
			Row: v1.SqlV1alpha1ResultItemRow{Items: []v1.SqlV1alpha1ResultItemRowOneOf{
				{SqlV1alpha1ResultItemString: &valAfter},
			}},
		})
	}

	return types.MockStatementResult{
		ResultSchema: v1.SqlV1alpha1ResultSchema{Columns: &columnDetails},
		StatementResults: v1.SqlV1alpha1StatementResult{
			Results: &v1.SqlV1alpha1StatementResultResults{Data: &resultData},
		},
	}
}

// TODO - This was only used for debugging/testing as gateway as broken
func ShowTablesSchema() v1.SqlV1alpha1ResultSchema {
	var columnDetails []v1.ColumnDetails
	dataType := v1.VarcharTypeAsDataType(&v1.VarcharType{
		Nullable: false,
		Type:     "VARCHAR",
	})
	columnDetails = append(columnDetails, v1.ColumnDetails{
		Name: "Table Name",
		Type: dataType,
	})

	return v1.SqlV1alpha1ResultSchema{Columns: &columnDetails}
}
