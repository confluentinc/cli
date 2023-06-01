package results

import (
	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

var nullField = types.AtomicStatementResultField{
	Type:  types.NULL,
	Value: "NULL",
}

type SDKToStatementResultFieldConverter func(any) types.StatementResultField

func GetConverterForType(dataType flinkgatewayv1alpha1.DataType) SDKToStatementResultFieldConverter {
	fieldType := types.NewResultFieldType(dataType)
	switch fieldType {
	case types.ARRAY:
		elementType := dataType.GetElementType()
		return toArrayStatementResultFieldConverter(elementType)
	case types.MULTISET:
		keyType := dataType.GetElementType()
		valueType := flinkgatewayv1alpha1.DataType{
			Nullable: false,
			Type:     "INTEGER",
		}
		return toMapStatementResultFieldConverter(fieldType, keyType, valueType)
	case types.MAP:
		keyType := dataType.GetKeyType()
		valueType := dataType.GetValueType()
		return toMapStatementResultFieldConverter(fieldType, keyType, valueType)
	case types.ROW:
		elementTypes := dataType.GetFields()
		return toRowStatementResultFieldConverter(elementTypes)
	default:
		return toAtomicStatementResultFieldConverter(fieldType)
	}
}

func toAtomicStatementResultFieldConverter(fieldType types.StatementResultFieldType) SDKToStatementResultFieldConverter {
	return func(field any) types.StatementResultField {
		atomicField, ok := field.(string)
		if !ok {
			return nullField
		}
		return types.AtomicStatementResultField{
			Type:  fieldType,
			Value: atomicField,
		}
	}
}

func toArrayStatementResultFieldConverter(elementType flinkgatewayv1alpha1.DataType) SDKToStatementResultFieldConverter {
	toStatementResultFieldConverter := GetConverterForType(elementType)
	return func(field any) types.StatementResultField {
		arrayField, ok := field.([]any)
		if !ok {
			return nullField
		}
		var values []types.StatementResultField
		for _, item := range arrayField {
			values = append(values, toStatementResultFieldConverter(item))
		}
		return types.ArrayStatementResultField{
			Type:        types.ARRAY,
			ElementType: types.NewResultFieldType(elementType),
			Values:      values,
		}
	}
}

func toMapStatementResultFieldConverter(fieldType types.StatementResultFieldType, keyType, valueType flinkgatewayv1alpha1.DataType) SDKToStatementResultFieldConverter {
	keyToStatementResultFieldConverter := GetConverterForType(keyType)
	valueToStatementResultFieldConverter := GetConverterForType(valueType)
	return func(field any) types.StatementResultField {
		mapField, ok := field.([]any)
		if !ok {
			return nullField
		}
		var entries []types.MapStatementResultFieldEntry
		for _, mapEntry := range mapField {
			mapEntry, ok := mapEntry.([]any)
			if !ok || len(mapEntry) != 2 {
				return nullField
			}

			key := mapEntry[0]
			value := mapEntry[1]
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

func toRowStatementResultFieldConverter(elementTypes []flinkgatewayv1alpha1.RowFieldType) SDKToStatementResultFieldConverter {
	return func(field any) types.StatementResultField {
		rowField, ok := field.([]any)
		if !ok || len(rowField) != len(elementTypes) {
			return nullField
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range rowField {
			elementType := elementTypes[idx].GetFieldType()
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
