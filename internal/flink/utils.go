package flink

import (
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

// copyDataType recursively converts the SDK DataType to a LocalDataType.
func copyDataType(sdkType cmfsdk.DataType) LocalDataType {
	localType := LocalDataType{
		Type:                sdkType.Type,
		Nullable:            sdkType.Nullable,
		Length:              sdkType.Length,
		Precision:           sdkType.Precision,
		Scale:               sdkType.Scale,
		Resolution:          sdkType.Resolution,
		FractionalPrecision: sdkType.FractionalPrecision,
	}
	if sdkType.KeyType != nil {
		copiedKeyType := copyDataType(*sdkType.KeyType)
		localType.KeyType = &copiedKeyType
	}
	if sdkType.ValueType != nil {
		copiedValueType := copyDataType(*sdkType.ValueType)
		localType.ValueType = &copiedValueType
	}
	if sdkType.ElementType != nil {
		copiedElementType := copyDataType(*sdkType.ElementType)
		localType.ElementType = &copiedElementType
	}
	if sdkType.Fields != nil {
		localFields := make([]LocalDataTypeField, 0, len(*sdkType.Fields))
		for _, sdkField := range *sdkType.Fields {
			localFields = append(localFields, LocalDataTypeField{
				Name:        sdkField.Name,
				FieldType:   copyDataType(sdkField.FieldType),
				Description: sdkField.Description,
			})
		}
		localType.Fields = &localFields
	}
	return localType
}
