package types

import (
	"fmt"
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"pgregory.net/rapid"
	"strconv"
)

func GetResultItemGeneratorForType(dataType v1.DataType) *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	fieldType := NewResultFieldType(dataType)
	switch fieldType {
	case ARRAY:
		elementType := dataType.ArrayType.GetArrayElementType()
		return ArrayResultItem(elementType)
	case MULTISET:
		keyType := dataType.MultisetType.GetMultisetElementType()
		valueType := v1.DataType{IntegerType: &v1.IntegerType{
			Nullable: false,
			Type:     "INTEGER",
		}}
		return MapResultItem(keyType, valueType)
	case MAP:
		keyType := dataType.MapType.GetKeyType()
		valueType := dataType.MapType.GetValueType()
		return MapResultItem(keyType, valueType)
	case ROW:
		elementTypes := dataType.RowType.GetFields()
		return RowResultItem(elementTypes)
	default:
		return AtomicResultItem()
	}
}

// AtomicResultItem generates a random atomic field
func AtomicResultItem() *rapid.Generator[v1.SqlV1alpha1ResultItemRowOneOf] {
	return rapid.Custom(func(t *rapid.T) v1.SqlV1alpha1ResultItemRowOneOf {
		atomicValue := v1.SqlV1alpha1ResultItemString(rapid.StringMatching("[a-zA-Z]+").Draw(t, "a string"))
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
		return v1.SqlV1alpha1ResultItem{
			Row: v1.SqlV1alpha1ResultItemRow{Items: items},
		}
	})
}

var NonAtomicResultFieldTypes = []StatementResultFieldType{
	ARRAY,
	MULTISET,
	MAP,
	ROW,
	NULL,
}

var AtomicResultFieldTypes = []StatementResultFieldType{
	CHAR,
	VARCHAR,
	BOOLEAN,
	BINARY,
	VARBINARY,
	DECIMAL,
	TINYINT,
	SMALLINT,
	INTEGER,
	BIGINT,
	FLOAT,
	DOUBLE,
	DATE,
	TIME_WITHOUT_TIME_ZONE,
	TIMESTAMP_WITHOUT_TIME_ZONE,
	TIMESTAMP_WITH_TIME_ZONE,
	TIMESTAMP_WITH_LOCAL_TIME_ZONE,
	INTERVAL_YEAR_MONTH,
	INTERVAL_DAY_TIME,
}

func getDataTypeGeneratorForType(fieldType StatementResultFieldType, maxNestingDepth int) *rapid.Generator[v1.DataType] {
	if maxNestingDepth <= 0 {
		return AtomicDataType()
	}
	switch fieldType {
	case ARRAY:
		return ArrayDataType(maxNestingDepth - 1)
	case MULTISET:
		return MultisetDataType(maxNestingDepth - 1)
	case MAP:
		return MapDataType(maxNestingDepth - 1)
	case ROW:
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

func GenResultFieldType() *rapid.Generator[StatementResultFieldType] {
	return rapid.Custom(func(t *rapid.T) StatementResultFieldType {
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

func MockResultColumns(numColumns int) *rapid.Generator[[]v1.ColumnDetails] {
	return rapid.Custom(func(t *rapid.T) []v1.ColumnDetails {
		var columnDetails []v1.ColumnDetails
		for i := 0; i < numColumns; i++ {
			maxNestingDepth := 5
			columnDetails = append(columnDetails, v1.ColumnDetails{
				Name: fmt.Sprintf("Column %v", i+1),
				Type: DataType(maxNestingDepth).Draw(t, "column type"),
			})
		}
		return columnDetails
	})
}

func MockResults(numColumns int) *rapid.Generator[MockStatementResult] {
	return rapid.Custom(func(t *rapid.T) MockStatementResult {
		if numColumns <= 0 {
			numColumns = rapid.IntRange(1, 10).Draw(t, "column number")
		}
		columnDetails := MockResultColumns(numColumns).Draw(t, "column details")
		resultData := rapid.SliceOfN(MockResultRow(columnDetails), 20, 50).Draw(t, "result data")
		return MockStatementResult{
			ResultSchema: v1.SqlV1alpha1ResultSchema{Columns: &columnDetails},
			StatementResults: v1.SqlV1alpha1StatementResult{
				Results: &v1.SqlV1alpha1StatementResultResults{Data: &resultData},
			},
		}
	})
}
