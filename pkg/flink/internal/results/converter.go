package results

import (
	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

var nullField = types.AtomicStatementResultField{
	Type:  types.Null,
	Value: "NULL",
}

type SDKToStatementResultFieldConverter func(any) types.StatementResultField

func GetConverterForType(dataType flinkgatewayv1.DataType) SDKToStatementResultFieldConverter {
	fieldType := types.NewResultFieldType(dataType.GetType())
	switch fieldType {
	case types.Array:
		elementType := dataType.GetElementType()
		return toArrayStatementResultFieldConverter(elementType)
	case types.Multiset:
		keyType := dataType.GetElementType()
		valueType := flinkgatewayv1.DataType{
			Nullable: false,
			Type:     "INTEGER",
		}
		return toMapStatementResultFieldConverter(fieldType, keyType, valueType)
	case types.Map:
		keyType := dataType.GetKeyType()
		valueType := dataType.GetValueType()
		return toMapStatementResultFieldConverter(fieldType, keyType, valueType)
	case types.Row:
		elementTypes := dataType.GetFields()
		return toRowStatementResultFieldConverter(elementTypes)
	case types.StructuredType:
		elementTypes := dataType.GetFields()
		return toStructuredStatementResultFieldConverter(elementTypes)
	default:
		return toAtomicStatementResultFieldConverter(fieldType)
	}
}

func GetConverterForTypeOnPrem(dataType cmfsdk.DataType) SDKToStatementResultFieldConverter {
	fieldType := types.NewResultFieldType(dataType.GetType())
	switch fieldType {
	case types.Array:
		elementType := dataType.GetElementType()
		return toArrayStatementResultFieldConverterOnPrem(elementType)
	case types.Multiset:
		keyType := dataType.GetElementType()
		valueType := cmfsdk.DataType{
			Nullable: false,
			Type:     "INTEGER",
		}
		return toMapStatementResultFieldConverterOnPrem(fieldType, keyType, valueType)
	case types.Map:
		keyType := dataType.GetKeyType()
		valueType := dataType.GetValueType()
		return toMapStatementResultFieldConverterOnPrem(fieldType, keyType, valueType)
	case types.Row:
		elementTypes := dataType.GetFields()
		return toRowStatementResultFieldConverterOnPrem(elementTypes)
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

func toArrayStatementResultFieldConverter(elementType flinkgatewayv1.DataType) SDKToStatementResultFieldConverter {
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
			Type:        types.Array,
			ElementType: types.NewResultFieldType(elementType.GetType()),
			Values:      values,
		}
	}
}

func toArrayStatementResultFieldConverterOnPrem(elementType cmfsdk.DataType) SDKToStatementResultFieldConverter {
	toStatementResultFieldConverter := GetConverterForTypeOnPrem(elementType)
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
			Type:        types.Array,
			ElementType: types.NewResultFieldType(elementType.GetType()),
			Values:      values,
		}
	}
}

func toMapStatementResultFieldConverter(fieldType types.StatementResultFieldType, keyType, valueType flinkgatewayv1.DataType) SDKToStatementResultFieldConverter {
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
			KeyType:   types.NewResultFieldType(keyType.GetType()),
			ValueType: types.NewResultFieldType(valueType.GetType()),
			Entries:   entries,
		}
	}
}

func toMapStatementResultFieldConverterOnPrem(fieldType types.StatementResultFieldType, keyType, valueType cmfsdk.DataType) SDKToStatementResultFieldConverter {
	keyToStatementResultFieldConverter := GetConverterForTypeOnPrem(keyType)
	valueToStatementResultFieldConverter := GetConverterForTypeOnPrem(valueType)
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
			KeyType:   types.NewResultFieldType(keyType.GetType()),
			ValueType: types.NewResultFieldType(valueType.GetType()),
			Entries:   entries,
		}
	}
}

func toRowStatementResultFieldConverter(elementTypes []flinkgatewayv1.RowFieldType) SDKToStatementResultFieldConverter {
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
			Type:         types.Row,
			ElementTypes: elementResultFieldTypes,
			Values:       values,
		}
	}
}

func toRowStatementResultFieldConverterOnPrem(elementTypes []cmfsdk.DataTypeField) SDKToStatementResultFieldConverter {
	return func(field any) types.StatementResultField {
		rowField, ok := field.([]any)
		if !ok || len(rowField) != len(elementTypes) {
			return nullField
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range rowField {
			elementType := elementTypes[idx].GetFieldType()
			toStatementResultFieldConverter := GetConverterForTypeOnPrem(elementType)
			convertedElement := toStatementResultFieldConverter(item)
			elementResultFieldTypes = append(elementResultFieldTypes, convertedElement.GetType())
			values = append(values, convertedElement)
		}
		return types.RowStatementResultField{
			Type:         types.Row,
			ElementTypes: elementResultFieldTypes,
			Values:       values,
		}
	}
}

func toStructuredStatementResultFieldConverter(fieldTypes []flinkgatewayv1.RowFieldType) SDKToStatementResultFieldConverter {
	return func(field any) types.StatementResultField {
		structuredField, ok := field.([]any)
		if !ok || len(structuredField) != len(fieldTypes) {
			return nullField
		}

		var elementNames []string
		var elementTypes []types.StatementResultFieldType
		var values []types.StatementResultField

		for idx, item := range structuredField {
			fieldSchema := fieldTypes[idx]
			elementName := fieldSchema.GetName()
			elementType := fieldSchema.GetFieldType()

			converter := GetConverterForType(elementType)
			converted := converter(item)

			elementNames = append(elementNames, elementName)
			elementTypes = append(elementTypes, converted.GetType())
			values = append(values, converted)
		}

		return types.StructuredTypeStatementResultField{
			Type:       types.StructuredType,
			FieldNames: elementNames,
			FieldTypes: elementTypes,
			Values:     values,
		}
	}
}
