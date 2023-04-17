package converter

import (
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/flink-sql-client/pkg/types"
)

var nullField = types.AtomicStatementResultField{
	Type:  types.NULL,
	Value: "NULL",
}

type SDKToStatementResultFieldConverter func(v1.SqlV1alpha1ResultItemRowOneOf) types.StatementResultField

func GetConverterForType(dataType v1.DataType) SDKToStatementResultFieldConverter {
	fieldType := types.NewResultFieldType(dataType)
	switch fieldType {
	case types.ARRAY:
		elementType := dataType.ArrayType.GetArrayElementType()
		return toArrayStatementResultFieldConverter(elementType)
	case types.MULTISET:
		keyType := dataType.MultisetType.GetMultisetElementType()
		valueType := v1.DataType{IntegerType: &v1.IntegerType{
			Nullable: false,
			Type:     "INTEGER",
		}}
		return toMapStatementResultFieldConverter(fieldType, keyType, valueType)
	case types.MAP:
		keyType := dataType.MapType.GetKeyType()
		valueType := dataType.MapType.GetValueType()
		return toMapStatementResultFieldConverter(fieldType, keyType, valueType)
	case types.ROW:
		elementTypes := dataType.RowType.GetFields()
		return toRowStatementResultFieldConverter(elementTypes)
	default:
		return toAtomicStatementResultFieldConverter(fieldType)
	}
}

func toAtomicStatementResultFieldConverter(fieldType types.StatementResultFieldType) SDKToStatementResultFieldConverter {
	return func(field v1.SqlV1alpha1ResultItemRowOneOf) types.StatementResultField {
		if field.SqlV1alpha1ResultItemString == nil {
			return nullField
		}
		return types.AtomicStatementResultField{
			Type:  fieldType,
			Value: string(*field.SqlV1alpha1ResultItemString),
		}
	}
}

func toArrayStatementResultFieldConverter(elementType v1.DataType) SDKToStatementResultFieldConverter {
	toStatementResultFieldConverter := GetConverterForType(elementType)
	return func(field v1.SqlV1alpha1ResultItemRowOneOf) types.StatementResultField {
		if field.SqlV1alpha1ResultItemRow == nil {
			return nullField
		}
		var values []types.StatementResultField
		for _, item := range field.SqlV1alpha1ResultItemRow.Items {
			values = append(values, toStatementResultFieldConverter(item))
		}
		return types.ArrayStatementResultField{
			Type:        types.ARRAY,
			ElementType: types.NewResultFieldType(elementType),
			Values:      values,
		}
	}
}

func toMapStatementResultFieldConverter(fieldType types.StatementResultFieldType, keyType, valueType v1.DataType) SDKToStatementResultFieldConverter {
	keyToStatementResultFieldConverter := GetConverterForType(keyType)
	valueToStatementResultFieldConverter := GetConverterForType(valueType)
	return func(field v1.SqlV1alpha1ResultItemRowOneOf) types.StatementResultField {
		if field.SqlV1alpha1ResultItemRow == nil {
			return nullField
		}
		var entries []types.MapStatementResultFieldEntry
		for _, mapEntry := range field.SqlV1alpha1ResultItemRow.Items {
			if mapEntry.SqlV1alpha1ResultItemRow == nil || len(mapEntry.SqlV1alpha1ResultItemRow.Items) != 2 {
				return nullField
			}

			key := mapEntry.SqlV1alpha1ResultItemRow.Items[0]
			value := mapEntry.SqlV1alpha1ResultItemRow.Items[1]
			entry := types.MapStatementResultFieldEntry{
				Key:   keyToStatementResultFieldConverter(key),
				Value: valueToStatementResultFieldConverter(value),
			}
			entries = append(entries, entry)
		}
		return types.MapStatementResultField{
			Type:      fieldType,
			KeyType:   types.NewResultFieldType(keyType),
			ValueType: types.NewResultFieldType(valueType),
			Entries:   entries,
		}
	}
}

func toRowStatementResultFieldConverter(elementTypes []v1.RowFieldType) SDKToStatementResultFieldConverter {
	return func(field v1.SqlV1alpha1ResultItemRowOneOf) types.StatementResultField {
		if field.SqlV1alpha1ResultItemRow == nil || len(field.SqlV1alpha1ResultItemRow.Items) != len(elementTypes) {
			return nullField
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range field.SqlV1alpha1ResultItemRow.Items {
			elementType := elementTypes[idx].GetType()
			toStatementResultFieldConverter := GetConverterForType(elementType)
			convertedElement := toStatementResultFieldConverter(item)
			elementResultFieldTypes = append(elementResultFieldTypes, convertedElement.GetType())
			values = append(values, convertedElement)
		}
		return types.RowStatementResultField{
			Type:         types.ROW,
			ElementTypes: elementResultFieldTypes,
			Values:       values,
		}
	}
}
