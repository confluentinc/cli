package results

import (
	"errors"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func convertToInternalField(field any, details flinkgatewayv1alpha1.ColumnDetails) types.StatementResultField {
	converter := GetConverterForType(details.GetType())
	if converter != nil {
		return converter(field)
	}

	return types.AtomicStatementResultField{
		Type:  types.NULL,
		Value: "NULL",
	}
}

func ConvertToInternalResults(results []any, resultSchema flinkgatewayv1alpha1.SqlV1alpha1ResultSchema) (*types.StatementResults, error) {
	header := make([]string, len(resultSchema.GetColumns()))
	for idx, column := range resultSchema.GetColumns() {
		header[idx] = column.GetName()
	}

	convertedResults := make([]types.StatementResultRow, len(results))
	for rowIdx, result := range results {
		resultItem, ok := result.(map[string]any)
		if !ok {
			return nil, errors.New("given result item does not match op/row schema")
		}

		items, _ := resultItem["row"].([]any)
		if len(items) != len(resultSchema.GetColumns()) {
			return nil, errors.New("given result row does not match the provided schema")
		}

		convertedFields := make([]types.StatementResultField, len(items))
		for colIdx, field := range items {
			columnSchema := resultSchema.GetColumns()[colIdx]
			convertedFields[colIdx] = convertToInternalField(field, columnSchema)
		}

		op, _ := resultItem["op"].(int32)
		convertedResults[rowIdx] = types.StatementResultRow{
			Operation: types.StatementResultOperation(op),
			Fields:    convertedFields,
		}
	}
	return &types.StatementResults{
		Headers: header,
		Rows:    convertedResults,
	}, nil
}
