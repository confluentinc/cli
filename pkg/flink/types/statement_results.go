package types

import (
	"strings"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

type StatementResults struct {
	Headers []string
	Rows    []StatementResultRow
}

func (s *StatementResults) GetHeaders() []string {
	if s == nil {
		return []string{}
	}
	return s.Headers
}

func (s *StatementResults) GetRows() []StatementResultRow {
	if s == nil {
		return []StatementResultRow{}
	}
	return s.Rows
}

type StatementResultRow struct {
	Operation StatementResultOperation
	Fields    []StatementResultField
}

func (r *StatementResultRow) GetRowKey(upsertColumns *[]int32) (string, bool) {
	rowValues := r.tryGetRowKeyWithUpsertColumns(upsertColumns)
	if len(rowValues) > 0 {
		return strings.Join(rowValues, "-"), true
	}

	// fallback to using the whole row as a key, in case we couldn't construct a row key with upsert columns
	for _, field := range r.GetFields() {
		rowValues = append(rowValues, field.ToString())
	}
	return strings.Join(rowValues, "-"), false
}

func (r *StatementResultRow) tryGetRowKeyWithUpsertColumns(upsertColumns *[]int32) []string {
	if upsertColumns == nil || len(*upsertColumns) == 0 {
		return nil
	}

	rowValues := []string{}
	for _, upsertColumnIdxInt32 := range *upsertColumns {
		upsertColumnIdx := int(upsertColumnIdxInt32)
		if upsertColumnIdx < 0 || upsertColumnIdx > len(r.GetFields()) {
			return nil
		}
		rowValues = append(rowValues, r.GetFields()[upsertColumnIdx].ToString())
	}
	return rowValues
}

func (r *StatementResultRow) GetFields() []StatementResultField {
	if r == nil {
		var fields []StatementResultField
		return fields
	}
	return r.Fields
}

const (
	Insert       StatementResultOperation = 0
	UpdateBefore StatementResultOperation = 1
	UpdateAfter  StatementResultOperation = 2
	Delete       StatementResultOperation = 3
)

type StatementResultOperation float64

func (s StatementResultOperation) IsInsertOperation() bool {
	return s == Insert || s == UpdateAfter
}

func (s StatementResultOperation) String() string {
	switch s {
	case Insert:
		return "+I"
	case UpdateBefore:
		return "-U"
	case UpdateAfter:
		return "+U"
	case Delete:
		return "-D"
	}
	return ""
}

type MockStatementResult struct {
	ResultSchema     flinkgatewayv1.SqlV1ResultSchema
	StatementResults flinkgatewayv1.SqlV1StatementResult
}

type MockStatementResultOnPrem struct {
	ResultSchema     cmfsdk.ResultSchema
	StatementResults cmfsdk.StatementResult
}
