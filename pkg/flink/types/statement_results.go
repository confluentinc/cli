package types

import (
	"strings"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
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

func (r *StatementResultRow) GetRowKey() string {
	rowKey := strings.Builder{}
	for idx, field := range r.GetFields() {
		rowKey.WriteString(field.ToString())
		if idx != len(r.GetFields())-1 {
			rowKey.WriteString("-")
		}
	}
	return rowKey.String()
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
