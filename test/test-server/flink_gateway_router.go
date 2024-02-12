package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
)

var flinkGatewayRoutes = []route{
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements", handleSqlEnvironmentsEnvironmentStatements},
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements/{statement}", handleSqlEnvironmentsEnvironmentStatementsStatement},
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements/{statement}/exceptions", handleSqlEnvironmentsEnvironmentStatementExceptions},
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
		switch r.Method {
		case http.MethodGet:
			statements := flinkgatewayv1.SqlV1StatementList{Data: []flinkgatewayv1.SqlV1Statement{{
				Name: flinkgatewayv1.PtrString("11111111-1111-1111-1"),
				Spec: &flinkgatewayv1.SqlV1StatementSpec{
					Statement:     flinkgatewayv1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1.PtrString("lfcp-123456"),
				},
				Status: &flinkgatewayv1.SqlV1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("SQL statement is completed"),
				},
				Metadata: &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))},
			}, {
				Name: flinkgatewayv1.PtrString("22222222-2222-2222-2"),
				Spec: &flinkgatewayv1.SqlV1StatementSpec{
					Statement:     flinkgatewayv1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1.PtrString("lfcp-123456"),
				},
				Status: &flinkgatewayv1.SqlV1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("SQL statement is completed"),
				},
				Metadata: &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))},
			}}}

			err := json.NewEncoder(w).Encode(statements)
			require.NoError(t, err)
		case http.MethodPost:
			statement := &flinkgatewayv1.SqlV1Statement{}
			err := json.NewDecoder(r.Body).Decode(statement)
			require.NoError(t, err)

			statement.Metadata = &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))}
			statement.Spec.ComputePoolId = flinkgatewayv1.PtrString("lfcp-123456")
			statement.Status = &flinkgatewayv1.SqlV1StatementStatus{Phase: "PENDING"}

			err = json.NewEncoder(w).Encode(statement)
			require.NoError(t, err)
		}
	}
}

func handleSqlEnvironmentsEnvironmentStatementExceptions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statement := flinkgatewayv1.SqlV1StatementExceptionList{
			Data: []flinkgatewayv1.SqlV1StatementException{{
				Timestamp: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				Name:      flinkgatewayv1.PtrString("Bad exception"),
				Message:   flinkgatewayv1.PtrString("exception in foo.go"),
			}, {
				Timestamp: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				Name:      flinkgatewayv1.PtrString("another Bad exception"),
				Message:   flinkgatewayv1.PtrString("exception in bar.go"),
			}},
		}

		err := json.NewEncoder(w).Encode(statement)
		require.NoError(t, err)
	}
}

func handleSqlEnvironmentsEnvironmentStatementsStatement(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statement := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(mux.Vars(r)["statement"]),
			Spec: &flinkgatewayv1.SqlV1StatementSpec{
				Statement:     flinkgatewayv1.PtrString("CREATE TABLE test;"),
				ComputePoolId: flinkgatewayv1.PtrString("pool-123456"),
			},
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: flinkgatewayv1.PtrString("SQL statement is completed"),
			},
			Metadata: &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))},
		}

		err := json.NewEncoder(w).Encode(statement)
		require.NoError(t, err)
	}
}
