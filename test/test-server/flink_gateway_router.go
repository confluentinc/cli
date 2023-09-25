package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"
)

var flinkGatewayRoutes = []route{
	{"/sql/v1beta1/organizations/{organization_id}/environments/{environment}/statements", handleSqlEnvironmentsEnvironmentStatements},
	{"/sql/v1beta1/organizations/{organization_id}/environments/{environment}/statements/{statement}", handleSqlEnvironmentsEnvironmentStatementsStatement},
	{"/sql/v1beta1/organizations/{organization_id}/environments/{environment}/statements/{statement}/exceptions", handleSqlEnvironmentsEnvironmentStatementExceptions},
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
		statements := flinkgatewayv1beta1.SqlV1beta1StatementList{Data: []flinkgatewayv1beta1.SqlV1beta1Statement{
			{
				Name: flinkgatewayv1beta1.PtrString("11111111-1111-1111-1"),
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement:     flinkgatewayv1beta1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1beta1.PtrString("lfcp-123456"),
				},
				Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1beta1.PtrString("SQL statement is completed"),
				},
				Metadata: &flinkgatewayv1beta1.ObjectMeta{CreatedAt: &creationDateStatement1},
			},
			{
				Name: flinkgatewayv1beta1.PtrString("22222222-2222-2222-2"),
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement:     flinkgatewayv1beta1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1beta1.PtrString("lfcp-123456"),
				},
				Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1beta1.PtrString("SQL statement is completed"),
				},
				Metadata: &flinkgatewayv1beta1.ObjectMeta{CreatedAt: &creationDateStatement2},
			},
		}}

		err := json.NewEncoder(w).Encode(statements)
		require.NoError(t, err)
	}
}

func handleSqlEnvironmentsEnvironmentStatementExceptions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statementExceptionCreatedAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		statementExceptionCreatedAt2 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		if r.Method == http.MethodGet {
			statement := flinkgatewayv1beta1.SqlV1beta1StatementExceptionList{
				Data: []flinkgatewayv1beta1.SqlV1beta1StatementException{{
					Timestamp:  &statementExceptionCreatedAt,
					Name:       flinkgatewayv1beta1.PtrString("Bad exception"),
					Stacktrace: flinkgatewayv1beta1.PtrString("exception in foo.go"),
				}, {
					Timestamp:  &statementExceptionCreatedAt2,
					Name:       flinkgatewayv1beta1.PtrString("another Bad exception"),
					Stacktrace: flinkgatewayv1beta1.PtrString("exception in bar.go"),
				}},
			}
			err := json.NewEncoder(w).Encode(statement)
			require.NoError(t, err)
		}
	}
}

func handleSqlEnvironmentsEnvironmentStatementsStatement(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statementCreatedAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		if r.Method == http.MethodGet {
			statement := flinkgatewayv1beta1.SqlV1beta1Statement{
				Name: flinkgatewayv1beta1.PtrString(mux.Vars(r)["statement"]),
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement:     flinkgatewayv1beta1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1beta1.PtrString("pool-123456"),
				},
				Status:   &flinkgatewayv1beta1.SqlV1beta1StatementStatus{Phase: "PENDING"},
				Metadata: &flinkgatewayv1beta1.ObjectMeta{CreatedAt: &statementCreatedAt},
			}
			err := json.NewEncoder(w).Encode(statement)
			require.NoError(t, err)
		}
	}
}
