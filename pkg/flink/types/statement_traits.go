package types

import (
	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

type StatementTraits struct {
	FlinkGatewayv1StatementTraits *flinkgatewayv1.SqlV1StatementTraits
	CmfStatementTraits            *cmfsdk.StatementTraits
}

func (s *StatementTraits) GetSqlKind() string {
	if s.FlinkGatewayv1StatementTraits != nil {
		return s.FlinkGatewayv1StatementTraits.GetSqlKind()
	} else if s.CmfStatementTraits != nil {
		return s.CmfStatementTraits.GetSqlKind()
	}
	return ""
}

func (s *StatementTraits) GetUpsertColumns() *[]int32 {
	if s.FlinkGatewayv1StatementTraits != nil {
		return s.FlinkGatewayv1StatementTraits.UpsertColumns
	} else if s.CmfStatementTraits != nil {
		return s.CmfStatementTraits.UpsertColumns
	}
	return nil
}

func (s *StatementTraits) GetColumnNames() []string {
	var columnNames []string
	if s.FlinkGatewayv1StatementTraits != nil {
		columns := s.FlinkGatewayv1StatementTraits.Schema.GetColumns()
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
