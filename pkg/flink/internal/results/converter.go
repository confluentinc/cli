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

type SDKToStatementResultFieldConverter func(any) (types.StatementResultField, error)

func GetConverterForType(dataType flinkgatewayv1.DataType) (SDKToStatementResultFieldConverter, error) {
	fieldType, err := types.NewResultFieldType(dataType.GetType())
	if err != nil {
		return nil, err
	}
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
		return toAtomicStatementResultFieldConverter(fieldType), nil
	}
}

func GetConverterForTypeOnPrem(dataType cmfsdk.DataType) (SDKToStatementResultFieldConverter, error) {
	fieldType, err := types.NewResultFieldType(dataType.GetType())
	if err != nil {
		return nil, err
	}
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
		return toAtomicStatementResultFieldConverter(fieldType), nil
	}
}

func toAtomicStatementResultFieldConverter(fieldType types.StatementResultFieldType) SDKToStatementResultFieldConverter {
	return func(field any) (types.StatementResultField, error) {
		atomicField, ok := field.(string)
		if !ok {
			return nullField, nil
		}
		return types.AtomicStatementResultField{
			Type:  fieldType,
			Value: atomicField,
		}, nil
	}
}

func toArrayStatementResultFieldConverter(elementType flinkgatewayv1.DataType) (SDKToStatementResultFieldConverter, error) {
	toStatementResultFieldConverter, err := GetConverterForType(elementType)
	if err != nil {
		return nil, err
	}
	resultElementType, err := types.NewResultFieldType(elementType.GetType())
	if err != nil {
		return nil, err
	}
	return func(field any) (types.StatementResultField, error) {
		arrayField, ok := field.([]any)
		if !ok {
			return nullField, nil
		}
		var values []types.StatementResultField
		for _, item := range arrayField {
			value, err := toStatementResultFieldConverter(item)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return types.ArrayStatementResultField{
			Type:        types.Array,
			ElementType: resultElementType,
			Values:      values,
		}, nil
	}, nil
}

func toArrayStatementResultFieldConverterOnPrem(elementType cmfsdk.DataType) (SDKToStatementResultFieldConverter, error) {
	toStatementResultFieldConverter, err := GetConverterForTypeOnPrem(elementType)
	if err != nil {
		return nil, err
	}
	resultElementType, err := types.NewResultFieldType(elementType.GetType())
	if err != nil {
		return nil, err
	}
	return func(field any) (types.StatementResultField, error) {
		arrayField, ok := field.([]any)
		if !ok {
			return nullField, nil
		}
		var values []types.StatementResultField
		for _, item := range arrayField {
			value, err := toStatementResultFieldConverter(item)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return types.ArrayStatementResultField{
			Type:        types.Array,
			ElementType: resultElementType,
			Values:      values,
		}, nil
	}, nil
}

func toMapStatementResultFieldConverter(fieldType types.StatementResultFieldType, keyType, valueType flinkgatewayv1.DataType) (SDKToStatementResultFieldConverter, error) {
	keyToStatementResultFieldConverter, err := GetConverterForType(keyType)
	if err != nil {
		return nil, err
	}
	valueToStatementResultFieldConverter, err := GetConverterForType(valueType)
	if err != nil {
		return nil, err
	}
	resultKeyType, err := types.NewResultFieldType(keyType.GetType())
	if err != nil {
		return nil, err
	}
	resultValueType, err := types.NewResultFieldType(valueType.GetType())
	if err != nil {
		return nil, err
	}
	return func(field any) (types.StatementResultField, error) {
		mapField, ok := field.([]any)
		if !ok {
			return nullField, nil
		}
		var entries []types.MapStatementResultFieldEntry
		for _, mapEntry := range mapField {
			mapEntry, ok := mapEntry.([]any)
			if !ok || len(mapEntry) != 2 {
				return nullField, nil
			}

			key, err := keyToStatementResultFieldConverter(mapEntry[0])
			if err != nil {
				return nil, err
			}
			value, err := valueToStatementResultFieldConverter(mapEntry[1])
			if err != nil {
				return nil, err
			}
			entries = append(entries, types.MapStatementResultFieldEntry{Key: key, Value: value})
		}
		return types.MapStatementResultField{
			Type:      fieldType,
			KeyType:   resultKeyType,
			ValueType: resultValueType,
			Entries:   entries,
		}, nil
	}, nil
}

func toMapStatementResultFieldConverterOnPrem(fieldType types.StatementResultFieldType, keyType, valueType cmfsdk.DataType) (SDKToStatementResultFieldConverter, error) {
	keyToStatementResultFieldConverter, err := GetConverterForTypeOnPrem(keyType)
	if err != nil {
		return nil, err
	}
	valueToStatementResultFieldConverter, err := GetConverterForTypeOnPrem(valueType)
	if err != nil {
		return nil, err
	}
	resultKeyType, err := types.NewResultFieldType(keyType.GetType())
	if err != nil {
		return nil, err
	}
	resultValueType, err := types.NewResultFieldType(valueType.GetType())
	if err != nil {
		return nil, err
	}
	return func(field any) (types.StatementResultField, error) {
		mapField, ok := field.([]any)
		if !ok {
			return nullField, nil
		}
		var entries []types.MapStatementResultFieldEntry
		for _, mapEntry := range mapField {
			mapEntry, ok := mapEntry.([]any)
			if !ok || len(mapEntry) != 2 {
				return nullField, nil
			}

			key, err := keyToStatementResultFieldConverter(mapEntry[0])
			if err != nil {
				return nil, err
			}
			value, err := valueToStatementResultFieldConverter(mapEntry[1])
			if err != nil {
				return nil, err
			}
			entries = append(entries, types.MapStatementResultFieldEntry{Key: key, Value: value})
		}
		return types.MapStatementResultField{
			Type:      fieldType,
			KeyType:   resultKeyType,
			ValueType: resultValueType,
			Entries:   entries,
		}, nil
	}, nil
}

func toRowStatementResultFieldConverter(elementTypes []flinkgatewayv1.RowFieldType) (SDKToStatementResultFieldConverter, error) {
	return func(field any) (types.StatementResultField, error) {
		rowField, ok := field.([]any)
		if !ok || len(rowField) != len(elementTypes) {
			return nullField, nil
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range rowField {
			elementType := elementTypes[idx].GetFieldType()
			toStatementResultFieldConverter, err := GetConverterForType(elementType)
			if err != nil {
				return nil, err
			}
			convertedElement, err := toStatementResultFieldConverter(item)
			if err != nil {
				return nil, err
			}
			elementResultFieldTypes = append(elementResultFieldTypes, convertedElement.GetType())
			values = append(values, convertedElement)
		}
		return types.RowStatementResultField{
			Type:         types.Row,
			ElementTypes: elementResultFieldTypes,
			Values:       values,
		}, nil
	}, nil
}

func toRowStatementResultFieldConverterOnPrem(elementTypes []cmfsdk.DataTypeField) (SDKToStatementResultFieldConverter, error) {
	return func(field any) (types.StatementResultField, error) {
		rowField, ok := field.([]any)
		if !ok || len(rowField) != len(elementTypes) {
			return nullField, nil
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range rowField {
			elementType := elementTypes[idx].GetFieldType()
			toStatementResultFieldConverter, err := GetConverterForTypeOnPrem(elementType)
			if err != nil {
				return nil, err
			}
			convertedElement, err := toStatementResultFieldConverter(item)
			if err != nil {
				return nil, err
			}
			elementResultFieldTypes = append(elementResultFieldTypes, convertedElement.GetType())
			values = append(values, convertedElement)
		}
		return types.RowStatementResultField{
			Type:         types.Row,
			ElementTypes: elementResultFieldTypes,
			Values:       values,
		}, nil
	}, nil
}

func toStructuredStatementResultFieldConverter(fieldTypes []flinkgatewayv1.RowFieldType) (SDKToStatementResultFieldConverter, error) {
	return func(field any) (types.StatementResultField, error) {
		structuredField, ok := field.([]any)
		if !ok || len(structuredField) != len(fieldTypes) {
			return nullField, nil
		}

		var elementNames []string
		var elementTypes []types.StatementResultFieldType
		var values []types.StatementResultField

		for idx, item := range structuredField {
			fieldSchema := fieldTypes[idx]
			elementName := fieldSchema.GetName()
			elementType := fieldSchema.GetFieldType()

			converter, err := GetConverterForType(elementType)
			if err != nil {
				return nil, err
			}
			converted, err := converter(item)
			if err != nil {
				return nil, err
			}

			elementNames = append(elementNames, elementName)
			elementTypes = append(elementTypes, converted.GetType())
			values = append(values, converted)
		}

		return types.StructuredTypeStatementResultField{
			Type:       types.StructuredType,
			FieldNames: elementNames,
			FieldTypes: elementTypes,
			Values:     values,
		}, nil
	}, nil
}
