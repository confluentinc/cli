package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"
)

var flinkGatewayRoutes = []route{
	{"/sql/v1alpha1/environments/{environment}/statements", handleSqlEnvironmentsEnvironmentStatements},
	{"/sql/v1alpha1/environments/{environment}/statements/{statement}", handleSqlEnvironmentsEnvironmentStatementsStatement},
}

func NewFlinkGatewayRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range flinkGatewayRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}

func handleSqlEnvironmentsEnvironmentStatements(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		creationDateStatement1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		creationDateStatement2 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		statements := flinkgatewayv1alpha1.SqlV1alpha1StatementList{Data: []flinkgatewayv1alpha1.SqlV1alpha1Statement{
			{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					StatementName: flinkgatewayv1alpha1.PtrString("11111111-1111-1111-1"),
					Statement:     flinkgatewayv1alpha1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1alpha1.PtrString("lfcp-123456"),
				},
				Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1alpha1.PtrString("SQL statement is completed"),
				},
				Metadata: &flinkgatewayv1alpha1.ObjectMeta{CreatedAt: &creationDateStatement1},
			},
			{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					StatementName: flinkgatewayv1alpha1.PtrString("22222222-2222-2222-2"),
					Statement:     flinkgatewayv1alpha1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1alpha1.PtrString("lfcp-123456"),
				},
				Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1alpha1.PtrString("SQL statement is completed"),
				},
				Metadata: &flinkgatewayv1alpha1.ObjectMeta{CreatedAt: &creationDateStatement2},
			},
		}}

		err := json.NewEncoder(w).Encode(statements)
		require.NoError(t, err)
	}
}

func handleSqlEnvironmentsEnvironmentStatementsStatement(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			statement := flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					StatementName: flinkgatewayv1alpha1.PtrString(mux.Vars(r)["statement"]),
					Statement:     flinkgatewayv1alpha1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1alpha1.PtrString("pool-123456"),
				},
				Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
					Phase:  "TODO",
					Detail: flinkgatewayv1alpha1.PtrString("TODO"),
				},
			}
			err := json.NewEncoder(w).Encode(statement)
			require.NoError(t, err)
		}
	}
}
