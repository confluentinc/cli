package converter

import (
	"errors"
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/flink-sql-client/pkg/types"
)

func convertToInternalField(field v1.SqlV1alpha1ResultItemRowOneOf, details v1.ColumnDetails) types.StatementResultField {
	converter := GetConverterForType(details.GetType())
	if converter != nil {
		return converter(field)
	}

	return types.AtomicStatementResultField{
		Type:  types.NULL,
		Value: "NULL",
	}
}

func ConvertToInternalResults(results []v1.SqlV1alpha1ResultItem, resultSchema v1.SqlV1alpha1ResultSchema) ([]types.StatementResultColumn, error) {
	var convertedResults []types.StatementResultColumn
	for _, column := range resultSchema.GetColumns() {
		convertedResults = append(convertedResults, types.StatementResultColumn{
			Name: column.GetName(),
			Type: types.NewResultFieldType(column.GetType()),
		})
	}

	for _, result := range results {
		row := result.GetRow()
		if len(row.Items) != len(resultSchema.GetColumns()) {
			return nil, errors.New("given result row does not match the provided schema")
		}
		for idx, field := range row.Items {
			columnSchema := resultSchema.GetColumns()[idx]
			convertedResults[idx].Fields = append(convertedResults[idx].Fields, convertToInternalField(field, columnSchema))
		}
	}
	return convertedResults, nil
}
