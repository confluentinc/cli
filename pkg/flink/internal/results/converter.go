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

// GetConverterForType returns a converter for dataType along with dataType's
// field-type enum, so callers that also need the enum don't re-run
// NewResultFieldType on the same wire type.
func GetConverterForType(dataType flinkgatewayv1.DataType) (SDKToStatementResultFieldConverter, types.StatementResultFieldType, error) {
	fieldType, err := types.NewResultFieldType(dataType.GetType())
	if err != nil {
		return nil, "", err
	}

	var converter SDKToStatementResultFieldConverter
	switch fieldType {
	case types.Array:
		converter, err = toArrayStatementResultFieldConverter(dataType.GetElementType())
	case types.Multiset:
		valueType := flinkgatewayv1.DataType{Nullable: false, Type: "INTEGER"}
		converter, err = toMapStatementResultFieldConverter(fieldType, dataType.GetElementType(), valueType)
	case types.Map:
		converter, err = toMapStatementResultFieldConverter(fieldType, dataType.GetKeyType(), dataType.GetValueType())
	case types.Row:
		converter = toRowStatementResultFieldConverter(dataType.GetFields())
	case types.StructuredType:
		converter = toStructuredStatementResultFieldConverter(dataType.GetFields())
	default:
		converter = toAtomicStatementResultFieldConverter(fieldType)
	}
	if err != nil {
		return nil, "", err
	}
	return converter, fieldType, nil
}

// GetConverterForTypeOnPrem is the on-prem (CMF) counterpart of GetConverterForType.
func GetConverterForTypeOnPrem(dataType cmfsdk.DataType) (SDKToStatementResultFieldConverter, types.StatementResultFieldType, error) {
	fieldType, err := types.NewResultFieldType(dataType.GetType())
	if err != nil {
		return nil, "", err
	}

	var converter SDKToStatementResultFieldConverter
	switch fieldType {
	case types.Array:
		converter, err = toArrayStatementResultFieldConverterOnPrem(dataType.GetElementType())
	case types.Multiset:
		valueType := cmfsdk.DataType{Nullable: false, Type: "INTEGER"}
		converter, err = toMapStatementResultFieldConverterOnPrem(fieldType, dataType.GetElementType(), valueType)
	case types.Map:
		converter, err = toMapStatementResultFieldConverterOnPrem(fieldType, dataType.GetKeyType(), dataType.GetValueType())
	case types.Row:
		converter = toRowStatementResultFieldConverterOnPrem(dataType.GetFields())
	case types.StructuredType:
		converter = toStructuredStatementResultFieldConverterOnPrem(dataType.GetFields())
	default:
		converter = toAtomicStatementResultFieldConverter(fieldType)
	}
	if err != nil {
		return nil, "", err
	}
	return converter, fieldType, nil
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
	toStatementResultFieldConverter, resultElementType, err := GetConverterForType(elementType)
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
	toStatementResultFieldConverter, resultElementType, err := GetConverterForTypeOnPrem(elementType)
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
	keyConverter, resultKeyType, err := GetConverterForType(keyType)
	if err != nil {
		return nil, err
	}
	valueConverter, resultValueType, err := GetConverterForType(valueType)
	if err != nil {
		return nil, err
	}
	return mapFieldConverter(fieldType, resultKeyType, resultValueType, keyConverter, valueConverter), nil
}

func toMapStatementResultFieldConverterOnPrem(fieldType types.StatementResultFieldType, keyType, valueType cmfsdk.DataType) (SDKToStatementResultFieldConverter, error) {
	keyConverter, resultKeyType, err := GetConverterForTypeOnPrem(keyType)
	if err != nil {
		return nil, err
	}
	valueConverter, resultValueType, err := GetConverterForTypeOnPrem(valueType)
	if err != nil {
		return nil, err
	}
	return mapFieldConverter(fieldType, resultKeyType, resultValueType, keyConverter, valueConverter), nil
}

// mapFieldConverter is the closure body shared by the cloud and on-prem MAP/
// MULTISET converters, once their key/value element converters are resolved.
func mapFieldConverter(fieldType, resultKeyType, resultValueType types.StatementResultFieldType, keyConverter, valueConverter SDKToStatementResultFieldConverter) SDKToStatementResultFieldConverter {
	return func(field any) (types.StatementResultField, error) {
		mapField, ok := field.([]any)
		if !ok {
			return nullField, nil
		}
		var entries []types.MapStatementResultFieldEntry
		for _, raw := range mapField {
			entry, ok, err := convertMapEntry(keyConverter, valueConverter, raw)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nullField, nil
			}
			entries = append(entries, entry)
		}
		return types.MapStatementResultField{
			Type:      fieldType,
			KeyType:   resultKeyType,
			ValueType: resultValueType,
			Entries:   entries,
		}, nil
	}
}

// convertMapEntry converts one raw [key, value] pair. ok is false (with no error)
// when the pair doesn't match the expected shape, mirroring the nullField fallback
// the other converters use for a value that doesn't fit its declared type.
func convertMapEntry(keyConverter, valueConverter SDKToStatementResultFieldConverter, raw any) (types.MapStatementResultFieldEntry, bool, error) {
	pair, ok := raw.([]any)
	if !ok || len(pair) != 2 {
		return types.MapStatementResultFieldEntry{}, false, nil
	}
	key, err := keyConverter(pair[0])
	if err != nil {
		return types.MapStatementResultFieldEntry{}, false, err
	}
	value, err := valueConverter(pair[1])
	if err != nil {
		return types.MapStatementResultFieldEntry{}, false, err
	}
	return types.MapStatementResultFieldEntry{Key: key, Value: value}, true, nil
}

func toRowStatementResultFieldConverter(elementTypes []flinkgatewayv1.RowFieldType) SDKToStatementResultFieldConverter {
	return func(field any) (types.StatementResultField, error) {
		rowField, ok := field.([]any)
		if !ok || len(rowField) != len(elementTypes) {
			return nullField, nil
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range rowField {
			elementType := elementTypes[idx].GetFieldType()
			toStatementResultFieldConverter, _, err := GetConverterForType(elementType)
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
	}
}

func toRowStatementResultFieldConverterOnPrem(elementTypes []cmfsdk.DataTypeField) SDKToStatementResultFieldConverter {
	return func(field any) (types.StatementResultField, error) {
		rowField, ok := field.([]any)
		if !ok || len(rowField) != len(elementTypes) {
			return nullField, nil
		}
		var elementResultFieldTypes []types.StatementResultFieldType
		var values []types.StatementResultField
		for idx, item := range rowField {
			elementType := elementTypes[idx].GetFieldType()
			toStatementResultFieldConverter, _, err := GetConverterForTypeOnPrem(elementType)
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
	}
}

func toStructuredStatementResultFieldConverter(fieldTypes []flinkgatewayv1.RowFieldType) SDKToStatementResultFieldConverter {
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

			converter, _, err := GetConverterForType(elementType)
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
	}
}

func toStructuredStatementResultFieldConverterOnPrem(fieldTypes []cmfsdk.DataTypeField) SDKToStatementResultFieldConverter {
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

			converter, _, err := GetConverterForTypeOnPrem(elementType)
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
	}
}
