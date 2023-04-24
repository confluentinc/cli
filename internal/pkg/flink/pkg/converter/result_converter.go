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

func ConvertToInternalResults(results []v1.SqlV1alpha1ResultItem, resultSchema v1.SqlV1alpha1ResultSchema) (*types.StatementResults, error) {
	var header []string
	for _, column := range resultSchema.GetColumns() {
		header = append(header, column.GetName())
	}

	var convertedResults []types.StatementResultRow
	for _, result := range results {
		row := result.GetRow()
		if len(row.Items) != len(resultSchema.GetColumns()) {
			return nil, errors.New("given result row does not match the provided schema")
		}

		var convertedFields []types.StatementResultField
		for idx, field := range row.Items {
			columnSchema := resultSchema.GetColumns()[idx]
			convertedFields = append(convertedFields, convertToInternalField(field, columnSchema))
		}
		convertedResults = append(convertedResults, types.StatementResultRow{
			Fields: convertedFields,
		})
	}
	return &types.StatementResults{
		Headers: header,
		Rows:    convertedResults,
	}, nil
}
