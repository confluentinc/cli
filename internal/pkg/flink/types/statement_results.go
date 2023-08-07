package types

import (
	"strings"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"
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
	INSERT        StatementResultOperation = 0
	UPDATE_BEFORE StatementResultOperation = 1
	UPDATE_AFTER  StatementResultOperation = 2
	DELETE        StatementResultOperation = 3
)

type StatementResultOperation float64

func (s StatementResultOperation) IsInsertOperation() bool {
	return s == INSERT || s == UPDATE_AFTER
}

func (s StatementResultOperation) String() string {
	switch s {
	case INSERT:
		return "+I"
	case UPDATE_BEFORE:
		return "-U"
	case UPDATE_AFTER:
		return "+U"
	case DELETE:
		return "-D"
	}
	return ""
}

type MockStatementResult struct {
	ResultSchema     flinkgatewayv1alpha1.SqlV1alpha1ResultSchema
	StatementResults flinkgatewayv1alpha1.SqlV1alpha1StatementResult
}
