package types

import (
	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

type StatementTraits struct {
	FlinkGatewayV1StatementTraits *flinkgatewayv1.SqlV1StatementTraits
	CmfStatementTraits            *cmfsdk.StatementTraits
}

func (s *StatementTraits) GetSqlKind() string {
	if s.FlinkGatewayV1StatementTraits != nil {
		return s.FlinkGatewayV1StatementTraits.GetSqlKind()
	} else if s.CmfStatementTraits != nil {
		return s.CmfStatementTraits.GetSqlKind()
	}
	return ""
}

func (s *StatementTraits) GetUpsertColumns() *[]int32 {
	if s.FlinkGatewayV1StatementTraits != nil {
		return s.FlinkGatewayV1StatementTraits.UpsertColumns
	} else if s.CmfStatementTraits != nil {
		return s.CmfStatementTraits.UpsertColumns
	}
	return nil
}

// GetIsBounded reports whether the gateway has told us the statement produces a
// finite result set. The second return value is false when the trait is absent,
// which happens before the statement leaves PENDING.
func (s *StatementTraits) GetIsBounded() (bool, bool) {
	if s.FlinkGatewayV1StatementTraits != nil && s.FlinkGatewayV1StatementTraits.IsBounded != nil {
		return s.FlinkGatewayV1StatementTraits.GetIsBounded(), true
	} else if s.CmfStatementTraits != nil && s.CmfStatementTraits.IsBounded != nil {
		return s.CmfStatementTraits.GetIsBounded(), true
	}
	return false, false
}

// GetIsAppendOnly reports whether the statement only ever emits insertions, in
// which case the changelog and the materialized table are the same thing. The
// second return value is false when the trait is absent.
func (s *StatementTraits) GetIsAppendOnly() (bool, bool) {
	if s.FlinkGatewayV1StatementTraits != nil && s.FlinkGatewayV1StatementTraits.IsAppendOnly != nil {
		return s.FlinkGatewayV1StatementTraits.GetIsAppendOnly(), true
	} else if s.CmfStatementTraits != nil && s.CmfStatementTraits.IsAppendOnly != nil {
		return s.CmfStatementTraits.GetIsAppendOnly(), true
	}
	return false, false
}

func (s *StatementTraits) GetColumnNames() []string {
	var columnNames []string
	if s.FlinkGatewayV1StatementTraits != nil {
		columns := s.FlinkGatewayV1StatementTraits.Schema.GetColumns()
		columnNames = make([]string, len(columns))
		for i, column := range columns {
			columnNames[i] = column.GetName()
		}
		return columnNames
	} else if s.CmfStatementTraits != nil {
		columns := s.CmfStatementTraits.Schema.GetColumns()
		columnNames = make([]string, len(columns))
		for i, column := range columns {
			columnNames[i] = column.GetName()
		}
	}
	return columnNames
}
