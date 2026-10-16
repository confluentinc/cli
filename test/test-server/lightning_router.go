package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	rtcesqlv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1alpha1"
)

var lightningRoutes = []route{
	{"/query/v1alpha1", handleLightningQuery},
}

func NewLightningRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range lightningRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}

// handleLightningQuery serves POST /query/v1alpha1. Scenarios are keyed off the SQL text; a query
// that matches none of them gets a 400.
func handleLightningQuery(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		// Without this header, a query too large or slow to answer inline becomes an asynchronous job,
		// which the CLI can't read, so every request must carry it.
		require.Equal(t, "true", r.Header.Get("X-Lightning-Sync-Only"))
		require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer"))

		var request rtcesqlv1alpha1.QueryV1alpha1QueryRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.NotEmpty(t, request.GetCatalogName())
		require.NotEmpty(t, request.GetDatabaseName())

		// Like the real endpoint, a query gets at most 200 rows, or fewer if it asks for fewer; asking for
		// no limit or a larger one gets 200, and the response doesn't say whether rows were left out.
		var maxResultRows *int32
		rowLimit := 200
		if options, ok := request.GetOptionsOk(); ok && options.MaxResultRows != nil {
			maxResultRows = options.MaxResultRows
			if limit := int(*maxResultRows); limit > 0 && limit < rowLimit {
				rowLimit = limit
			}
		}

		var columns []map[string]any
		var rows [][]any

		// The real endpoint trims the statement too, so a SQL file's trailing newline doesn't matter.
		switch strings.TrimSpace(request.GetQuery()) {
		case "SELECT order_id, status FROM orders LIMIT 2":
			columns = []map[string]any{lightningColumn("order_id", "INTEGER"), lightningColumn("status", "VARCHAR")}
			rows = [][]any{{"1021", "SHIPPED"}, {"1044", "PENDING"}}
		case "SELECT status, order_id FROM orders LIMIT 2":
			// Columns deliberately not in alphabetical order: -o json/-o yaml must keep the SELECT order.
			columns = []map[string]any{lightningColumn("status", "VARCHAR"), lightningColumn("order_id", "INTEGER")}
			rows = [][]any{{"SHIPPED", "1021"}, {"PENDING", "1044"}}
		case "SELECT catalog_name, database_name FROM whoami":
			// Echoes the environment and cluster the query was sent to.
			columns = []map[string]any{lightningColumn("catalog_name", "VARCHAR"), lightningColumn("database_name", "VARCHAR")}
			rows = [][]any{{request.GetCatalogName(), request.GetDatabaseName()}}
		case "SELECT max_result_rows FROM request_options":
			// Echoes the row limit the query asked for, or NULL when it asked for none.
			columns = []map[string]any{lightningColumn("max_result_rows", "INTEGER")}
			var value any
			if maxResultRows != nil {
				value = strconv.Itoa(int(*maxResultRows))
			}
			rows = [][]any{{value}}
		case "SELECT * FROM scalar_types":
			columns = []map[string]any{
				lightningColumn("c_boolean", "BOOLEAN"),
				lightningColumn("c_tinyint", "TINYINT"),
				lightningColumn("c_smallint", "SMALLINT"),
				lightningColumn("c_integer", "INTEGER"),
				lightningColumn("c_bigint", "BIGINT"),
				lightningColumn("c_float", "FLOAT"),
				lightningColumn("c_double", "DOUBLE"),
				lightningColumn("c_decimal", "DECIMAL"),
				lightningColumn("c_varchar", "VARCHAR"),
				lightningColumn("c_varbinary", "VARBINARY"),
				lightningColumn("c_date", "DATE"),
				lightningColumn("c_time", "TIME_WITHOUT_TIME_ZONE"),
				lightningColumn("c_timestamp", "TIMESTAMP_WITHOUT_TIME_ZONE"),
				lightningColumn("c_timestamp_ltz", "TIMESTAMP_WITH_TIME_ZONE"),
				lightningColumn("c_null", "VARCHAR"),
			}
			rows = [][]any{{"TRUE", "-7", "300", "70000", "9007199254740993", "17.45", "3.14159", "149.00", "hello", "x'7f0203'", "2026-10-12", "13:45:30.123", "2026-10-12 13:45:30.123456", "2026-10-12 20:45:30.123456", nil}}
		case "SELECT * FROM complex_types":
			columns = []map[string]any{
				{"name": "c_array", "type": map[string]any{"type": "ARRAY", "nullable": true, "element_type": map[string]any{"type": "VARCHAR", "nullable": true}}},
				{"name": "c_map", "type": map[string]any{"type": "MAP", "nullable": true, "key_type": map[string]any{"type": "VARCHAR", "nullable": false}, "value_type": map[string]any{"type": "INTEGER", "nullable": true}}},
				{"name": "c_int_map", "type": map[string]any{"type": "MAP", "nullable": true, "key_type": map[string]any{"type": "INTEGER", "nullable": false}, "value_type": map[string]any{"type": "VARCHAR", "nullable": true}}},
				{"name": "c_multiset", "type": map[string]any{"type": "MULTISET", "nullable": true, "element_type": map[string]any{"type": "VARCHAR", "nullable": true}}},
				{"name": "c_row", "type": map[string]any{"type": "ROW", "nullable": true, "fields": []any{
					map[string]any{"name": "USER_ID", "field_type": map[string]any{"type": "VARCHAR", "nullable": false}},
					map[string]any{"name": "AGE", "field_type": map[string]any{"type": "INTEGER", "nullable": false}},
				}}},
				{"name": "c_empty_array", "type": map[string]any{"type": "ARRAY", "nullable": true, "element_type": map[string]any{"type": "INTEGER", "nullable": true}}},
				{"name": "c_null_array", "type": map[string]any{"type": "ARRAY", "nullable": true, "element_type": map[string]any{"type": "INTEGER", "nullable": true}}},
			}
			rows = [][]any{{
				[]any{"a", nil, "c"},
				[]any{[]any{"latency", "42"}, []any{"errors", "0"}},
				[]any{[]any{"1", "one"}, []any{"2", "two"}},
				[]any{[]any{"x", "2"}},
				[]any{"u-1", "42"},
				[]any{},
				nil,
			}}
		case "SELECT note FROM escapes":
			columns = []map[string]any{lightningColumn("note", "VARCHAR")}
			rows = [][]any{{"\x1b[31mred\x1b[0m\nfake row"}}
		case "SELECT id FROM hundreds":
			columns = []map[string]any{lightningColumn("id", "INTEGER")}
			rows = lightningIdRows(150)
		case "SELECT id FROM exactly_hundred":
			columns = []map[string]any{lightningColumn("id", "INTEGER")}
			rows = lightningIdRows(100)
		case "SELECT id FROM two_hundred_one":
			columns = []map[string]any{lightningColumn("id", "INTEGER")}
			rows = lightningIdRows(201)
		case "SELECT id FROM empty":
			columns = []map[string]any{lightningColumn("id", "INTEGER")}
			rows = [][]any{}
		case "SELECT id FROM unrecognized_column_type":
			columns = []map[string]any{lightningColumn("id", "NOT_A_REAL_TYPE")}
			rows = [][]any{{"1"}}
		case "SELECT id FROM mismatched_row":
			columns = []map[string]any{lightningColumn("id", "INTEGER")}
			rows = [][]any{{"1", "2"}}
		case "SELECT * FROM huge":
			writeLightningError(t, w, http.StatusUnprocessableEntity, "FE_ASYNC_REQUIRED", "Query requires asynchronous execution, which the sync-only request disallows: Query would be promoted to asynchronous execution.")
			return
		case "SELECT * FROM promoted":
			// A server that ignored X-Lightning-Sync-Only and accepted the query as an asynchronous job.
			w.Header().Set("Location", fmt.Sprintf("/query/v1alpha1/jobs/%s/%s/promoted/statement_1b2c3d4e5f6a/status", request.GetCatalogName(), request.GetDatabaseName()))
			w.WriteHeader(http.StatusAccepted)
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"api_version": "query/v1alpha1",
				"kind":        "Statement",
				"result": map[string]any{
					"statement_id":     "statement_1b2c3d4e5f6a",
					"environment_id":   request.GetCatalogName(),
					"kafka_cluster_id": request.GetDatabaseName(),
					"topic_name":       "promoted",
				},
			}))
			return
		case "SELECT * FROM missing":
			writeLightningError(t, w, http.StatusBadRequest, "FE_SQL_VALIDATION_ERROR", fmt.Sprintf("Table 'missing' not found in database '%s'.", request.GetDatabaseName()))
			return
		case "SELECT * FROM server_error":
			// Not retried, unlike the gateway errors.
			writeLightningError(t, w, http.StatusInternalServerError, "FE_INTERNAL_ERROR", "The statement could not be completed. Retry, and contact support with the request id if it persists.")
			return
		default:
			writeLightningError(t, w, http.StatusBadRequest, "FE_SQL_VALIDATION_ERROR", fmt.Sprintf("Unexpected query %q.", request.GetQuery()))
			return
		}

		if len(rows) > rowLimit {
			rows = rows[:rowLimit]
		}

		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"api_version": "query/v1alpha1",
			"kind":        "Statement",
			"result": map[string]any{
				"result_format": "JSON",
				"schema":        map[string]any{"columns": columns},
				"data":          rows,
			},
		}))
	}
}

func lightningColumn(name, columnType string) map[string]any {
	return map[string]any{"name": name, "type": map[string]any{"type": columnType, "nullable": true}}
}

func lightningIdRows(count int) [][]any {
	rows := make([][]any, count)
	for i := range rows {
		rows[i] = []any{strconv.Itoa(i + 1)}
	}
	return rows
}

func writeLightningError(t *testing.T, w http.ResponseWriter, status int, code, detail string) {
	const requestId = "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"
	w.Header().Set("X-Request-Id", requestId)
	w.WriteHeader(status)
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
		"errors": []map[string]any{{
			"id":     requestId,
			"status": strconv.Itoa(status),
			"code":   code,
			"title":  http.StatusText(status),
			"detail": detail,
		}},
	}))
}
