package results

import (
	"errors"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
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
	header := make([]string, len(resultSchema.GetColumns()))
	for idx, column := range resultSchema.GetColumns() {
		header[idx] = column.GetName()
	}

	convertedResults := make([]types.StatementResultRow, len(results))
	for rowIdx, result := range results {
		row := result.GetRow()
		if len(row.Items) != len(resultSchema.GetColumns()) {
			return nil, errors.New("given result row does not match the provided schema")
		}

		convertedFields := make([]types.StatementResultField, len(row.Items))
		for colIdx, field := range row.Items {
			columnSchema := resultSchema.GetColumns()[colIdx]
			convertedFields[colIdx] = convertToInternalField(field, columnSchema)
		}
		convertedResults[rowIdx] = types.StatementResultRow{
			Operation: types.StatementResultOperation(int8(result.GetOp())),
			Fields:    convertedFields,
		}
	}
	return &types.StatementResults{
		Headers: header,
		Rows:    convertedResults,
	}, nil
}
