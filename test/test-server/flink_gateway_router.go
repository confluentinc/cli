package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
)

const (
	validFlinkStatementPrincipalId   = "u-123456"
	validFlinkStatementComputePoolId = "lfcp-123456"
)

var flinkGatewayRoutes = []route{
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements", handleSqlEnvironmentsEnvironmentStatements},
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements/{statement}", handleSqlEnvironmentsEnvironmentStatementsStatement},
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements/{statement}/exceptions", handleSqlEnvironmentsEnvironmentStatementExceptions},
	{"/sql/v1/organizations/{organization_id}/environments/{environment}/statements/{statement}/results", handleSqlEnvironmentsEnvironmentStatementResults},
	{"/sql/v1/organizations/{organization_id}/environments/{environment_id}/connections", handleSqlEnvironmentsEnvironmentConnections},
	{"/sql/v1/organizations/{organization_id}/environments/{environment_id}/connections/{connection}", handleSqlEnvironmentsEnvironmentConnectionsConnection},
}

func NewFlinkGatewayRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range flinkGatewayRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}

func handleSqlEnvironmentsEnvironmentConnections(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			connections := flinkgatewayv1.SqlV1ConnectionList{Data: []flinkgatewayv1.SqlV1Connection{{
				Name: flinkgatewayv1.PtrString("11111111-1111-1111-1"),
				Spec: &flinkgatewayv1.SqlV1ConnectionSpec{
					ConnectionType: flinkgatewayv1.PtrString("OPENAI"),
					Endpoint:       flinkgatewayv1.PtrString("https://api.openai.com/v1/chat/completions"),
					AuthData: &flinkgatewayv1.SqlV1ConnectionSpecAuthDataOneOf{
						SqlV1PlaintextProvider: &flinkgatewayv1.SqlV1PlaintextProvider{
							Kind: lo.ToPtr("PlaintextProvider"),
							Data: lo.ToPtr("<REDACTED>"),
						}},
				},
				Status: &flinkgatewayv1.SqlV1ConnectionStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("Connection1 is completed"),
				},
				Metadata: &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))},
			}, {
				Name: flinkgatewayv1.PtrString("22222222-2222-2222-2"),
				Spec: &flinkgatewayv1.SqlV1ConnectionSpec{
					ConnectionType: flinkgatewayv1.PtrString("OPENAI"),
					Endpoint:       flinkgatewayv1.PtrString("https://api.openai.com/v1/chat/completions"),
					AuthData: &flinkgatewayv1.SqlV1ConnectionSpecAuthDataOneOf{
						SqlV1PlaintextProvider: &flinkgatewayv1.SqlV1PlaintextProvider{
							Kind: lo.ToPtr("PlaintextProvider"),
							Data: lo.ToPtr("<REDACTED>"),
						}},
				},
				Status: &flinkgatewayv1.SqlV1ConnectionStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("Connection2 is completed"),
				},
				Metadata: &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))},
			}}}
			setPageToken(&connections, &connections.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(connections)
			require.NoError(t, err)
		case http.MethodPost:
			connection := &flinkgatewayv1.SqlV1Connection{}
			err := json.NewDecoder(r.Body).Decode(connection)
			require.NoError(t, err)

			connection.Metadata = &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))}
			connection.Status = &flinkgatewayv1.SqlV1ConnectionStatus{Phase: "PENDING"}
			connection.Spec.AuthData.SqlV1PlaintextProvider.Data = lo.ToPtr("<REDACTED>")

			err = json.NewEncoder(w).Encode(connection)
			require.NoError(t, err)
		}
	}
}

func handleSqlEnvironmentsEnvironmentConnectionsConnection(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			connectionName := mux.Vars(r)["connection"]
			if strings.Contains(connectionName, "nonexist") {
				err := writeResourceNotFoundError(w)
				require.NoError(t, err)
				return
			}
			connection := flinkgatewayv1.SqlV1Connection{
				Name: flinkgatewayv1.PtrString("11111111-1111-1111-1"),
				Spec: &flinkgatewayv1.SqlV1ConnectionSpec{
					ConnectionType: flinkgatewayv1.PtrString("OPENAI"),
					Endpoint:       flinkgatewayv1.PtrString("https://api.openai.com/v1/chat/completions"),
					AuthData: &flinkgatewayv1.SqlV1ConnectionSpecAuthDataOneOf{
						SqlV1PlaintextProvider: &flinkgatewayv1.SqlV1PlaintextProvider{
							Kind: lo.ToPtr("PlaintextProvider"),
							Data: lo.ToPtr("<REDACTED>"),
						}},
				},
				Status: &flinkgatewayv1.SqlV1ConnectionStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("Connection1 is completed"),
				},
				Metadata: &flinkgatewayv1.ObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))},
			}
			err := json.NewEncoder(w).Encode(connection)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for "/sql/v1/organizations/{organization_id}/environments/{environment_id}/statements"
func handleSqlEnvironmentsEnvironmentStatements(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			statements := flinkgatewayv1.SqlV1StatementList{Data: []flinkgatewayv1.SqlV1Statement{{
				Name: flinkgatewayv1.PtrString("11111111-1111-1111-1"),
				Spec: &flinkgatewayv1.SqlV1StatementSpec{
					Statement:     flinkgatewayv1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1.PtrString(validFlinkStatementComputePoolId),
				},
				Status: &flinkgatewayv1.SqlV1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("SQL statement is completed"),
					LatestOffsets: &map[string]string{
						"customers_source": "partition:0,offset:9223372036854775808",
					},
					LatestOffsetsTimestamp: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
				Metadata: &flinkgatewayv1.StatementObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))},
			}, {
				Name: flinkgatewayv1.PtrString("22222222-2222-2222-2"),
				Spec: &flinkgatewayv1.SqlV1StatementSpec{
					Statement:     flinkgatewayv1.PtrString("CREATE TABLE test;"),
					ComputePoolId: flinkgatewayv1.PtrString(validFlinkStatementComputePoolId),
				},
				Status: &flinkgatewayv1.SqlV1StatementStatus{
					Phase:  "COMPLETED",
					Detail: flinkgatewayv1.PtrString("SQL statement is completed"),
					LatestOffsets: &map[string]string{
						"customers_source": "partition:0,offset:9223372036854775808",
					},
					LatestOffsetsTimestamp: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
				Metadata: &flinkgatewayv1.StatementObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))},
			}}}
			setPageToken(&statements, &statements.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(statements)
			require.NoError(t, err)
		case http.MethodPost:
			statement := &flinkgatewayv1.SqlV1Statement{}
			err := json.NewDecoder(r.Body).Decode(statement)
			require.NoError(t, err)

			statement.Metadata = &flinkgatewayv1.StatementObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))}
			statement.Spec.ComputePoolId = flinkgatewayv1.PtrString(validFlinkStatementComputePoolId)
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

// Handler for "/sql/v1/organizations/{organization_id}/environments/{environment_id}/statements/{statement_name}"
func handleSqlEnvironmentsEnvironmentStatementsStatement(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleStatementGet(t)(w, r)
		case http.MethodPut:
			handleStatementUpdate(t)(w, r)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}

func handleStatementGet(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statement := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(mux.Vars(r)["statement"]),
			Spec: &flinkgatewayv1.SqlV1StatementSpec{
				Statement: flinkgatewayv1.PtrString("CREATE TABLE test;"),
				Properties: &map[string]string{
					"sql.current-catalog":  "default",
					"sql.current-database": "my-cluster",
				},
				ComputePoolId: flinkgatewayv1.PtrString(validFlinkStatementComputePoolId),
				Principal:     flinkgatewayv1.PtrString(validFlinkStatementPrincipalId),
			},
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: flinkgatewayv1.PtrString("SQL statement is completed"),
				Traits: &flinkgatewayv1.SqlV1StatementTraits{
					Schema: &flinkgatewayv1.SqlV1ResultSchema{
						Columns: &[]flinkgatewayv1.ColumnDetails{
							{Name: "database_name", Type: flinkgatewayv1.DataType{Type: "VARCHAR"}},
						},
					},
				},
				LatestOffsets: &map[string]string{
					"customers_source": "partition:0,offset:9223372036854775808",
				},
				LatestOffsetsTimestamp: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			Metadata: &flinkgatewayv1.StatementObjectMeta{CreatedAt: flinkgatewayv1.PtrTime(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))},
		}

		err := json.NewEncoder(w).Encode(statement)
		require.NoError(t, err)
	}
}

func handleSqlEnvironmentsEnvironmentStatementResults(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := flinkgatewayv1.SqlV1StatementResult{
			Metadata: flinkgatewayv1.ResultListMeta{},
			Results: &flinkgatewayv1.SqlV1StatementResultResults{
				Data: &[]any{
					map[string]any{"op": float64(0), "row": []any{"my-cluster"}},
					map[string]any{"op": float64(0), "row": []any{"other-cluster"}},
				},
			},
		}

		err := json.NewEncoder(w).Encode(result)
		require.NoError(t, err)
	}
}

func handleStatementUpdate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(flinkgatewayv1.SqlV1Statement)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		stopped := req.Spec.GetStopped()
		principal := req.Spec.GetPrincipal()
		computePool := req.Spec.GetComputePoolId()

		// Handle the stop case, principal and computerPool shouldn't matter
		if stopped {
			w.WriteHeader(http.StatusAccepted)
			err = json.NewEncoder(w).Encode(flinkgatewayv1.NewSqlV1Statement())
			require.NoError(t, err)
			return
		}

		// Handle the resume case: invalid principal ID
		if principal != "" && principal != validFlinkStatementPrincipalId {
			w.WriteHeader(http.StatusBadRequest)
			err = writeError(w, "Bad Request")
			require.NoError(t, err)
			return
		}

		// Handle the resume case: invalid compute pool
		if computePool != "" && computePool != validFlinkStatementComputePoolId {
			w.WriteHeader(http.StatusBadRequest)
			err = writeError(w, fmt.Sprintf("logical compute pool=%s not found", computePool))
			require.NoError(t, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
