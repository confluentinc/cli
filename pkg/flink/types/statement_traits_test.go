package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func TestGetSqlKind(t *testing.T) {
	flinkGatewayStatementTraits := StatementTraits{FlinkGatewayV1StatementTraits: &flinkgatewayv1.SqlV1StatementTraits{
		SqlKind: flinkgatewayv1.PtrString("SELECT"),
	}}

	require.Equal(t, "SELECT", flinkGatewayStatementTraits.GetSqlKind())

	cmfStatementTraits := StatementTraits{CmfStatementTraits: &cmfsdk.StatementTraits{
		SqlKind: cmfsdk.PtrString("SELECT"),
	}}

	require.Equal(t, "SELECT", cmfStatementTraits.GetSqlKind())

	emptyStatementTraits := StatementTraits{}
	require.Equal(t, "", emptyStatementTraits.GetSqlKind())
}

func TestGetUpsertColumns(t *testing.T) {
	flinkGatewayStatementTraits := StatementTraits{FlinkGatewayV1StatementTraits: &flinkgatewayv1.SqlV1StatementTraits{
		UpsertColumns: &[]int32{0, 1},
	}}

	require.NotNil(t, flinkGatewayStatementTraits.GetUpsertColumns())
	require.Equal(t, []int32{0, 1}, *flinkGatewayStatementTraits.GetUpsertColumns())

	cmfStatementTraits := StatementTraits{CmfStatementTraits: &cmfsdk.StatementTraits{
		UpsertColumns: &[]int32{0, 1},
	}}

	require.NotNil(t, cmfStatementTraits.GetUpsertColumns())
	require.Equal(t, []int32{0, 1}, *cmfStatementTraits.GetUpsertColumns())
}

func TestGetColumnNames(t *testing.T) {
	flinkGatewayStatementTraits := StatementTraits{FlinkGatewayV1StatementTraits: &flinkgatewayv1.SqlV1StatementTraits{
		Schema: &flinkgatewayv1.SqlV1ResultSchema{
			Columns: &[]flinkgatewayv1.ColumnDetails{
				{Name: "column1"},
				{Name: "column2"},
			},
		},
	}}

	require.Equal(t, []string{"column1", "column2"}, flinkGatewayStatementTraits.GetColumnNames())

	cmfStatementTraits := StatementTraits{CmfStatementTraits: &cmfsdk.StatementTraits{
		Schema: &cmfsdk.ResultSchema{
			Columns: []cmfsdk.ResultSchemaColumn{
				{Name: "column1"},
				{Name: "column2"},
			},
		},
	}}

	require.Equal(t, []string{"column1", "column2"}, cmfStatementTraits.GetColumnNames())
}
