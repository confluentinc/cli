package types

import v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

type StatementResultColumn struct {
	Name   string
	Type   StatementResultFieldType
	Fields []StatementResultField
}

type MockStatementResult struct {
	ResultSchema     v1.SqlV1alpha1ResultSchema
	StatementResults v1.SqlV1alpha1StatementResult
}
