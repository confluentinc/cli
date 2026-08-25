package query

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	climock "github.com/confluentinc/cli/v4/mock"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	cliconfig "github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

func TestNew(t *testing.T) {
	cfg := cliconfig.AuthenticatedCloudConfigMock()
	prerunner := climock.NewPreRunnerMock(nil, nil, nil, nil, cfg)

	cmd := New(cfg, prerunner)

	require.Equal(t, "query [name]", cmd.Use)
	require.False(t, cmd.Hidden, "cfg.IsTest should keep the command visible in tests")

	for _, name := range []string{"sql", "compute-pool", "service-account", "database", "property", "timeout", "max-rows", "raw", "environment", "context", "output", "cloud", "region"} {
		require.NotNil(t, cmd.Flags().Lookup(name), "expected --%s to be registered", name)
	}

	sqlFlag := cmd.Flags().Lookup("sql")
	require.Equal(t, "true", sqlFlag.Annotations[cobra.BashCompOneRequiredFlag][0])
}

// captureStdout redirects the package-level os.Stdout (which output.Print and
// tablewriter both write to directly) for the duration of fn and returns what
// was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func newTestCommand(ctx *cliconfig.Context) *command {
	return &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{Context: ctx}}
}

func newTestContext(platformServer, authToken string) *cliconfig.Context {
	return &cliconfig.Context{
		Platform:  &cliconfig.Platform{Server: platformServer},
		State:     &cliconfig.ContextState{AuthToken: authToken},
		LastOrgId: "org-1",
	}
}

func TestBuildQueryProperties(t *testing.T) {
	tests := []struct {
		name     string
		catalog  string
		database string
		flags    []string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:    "catalog and default snapshot mode, no database",
			catalog: "env-123",
			expected: map[string]string{
				"sql.current-catalog": "env-123",
				"sql.snapshot.mode":   "now",
			},
		},
		{
			name:     "database is included when set",
			catalog:  "env-123",
			database: "my-cluster",
			expected: map[string]string{
				"sql.current-catalog":  "env-123",
				"sql.snapshot.mode":    "now",
				"sql.current-database": "my-cluster",
			},
		},
		{
			name:    "property flag overrides the snapshot mode default",
			catalog: "env-123",
			flags:   []string{"sql.snapshot.mode=earliest"},
			expected: map[string]string{
				"sql.current-catalog": "env-123",
				"sql.snapshot.mode":   "earliest",
			},
		},
		{
			name:    "malformed property flag is rejected",
			catalog: "env-123",
			flags:   []string{"not-a-key-value-pair"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().StringSlice("property", []string{}, "")
			for _, f := range tt.flags {
				require.NoError(t, cmd.Flags().Set("property", f))
			}

			c := newTestCommand(nil)
			got, err := c.buildQueryProperties(cmd, tt.catalog, tt.database)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func newOutputCmd(t *testing.T, format string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	pcmd.AddOutputFlag(cmd)
	if format != "" {
		require.NoError(t, cmd.Flags().Set("output", format))
	}
	return cmd
}

func testColumns() []flinkgatewayv1.ColumnDetails {
	return []flinkgatewayv1.ColumnDetails{
		{Name: "id", Type: flinkgatewayv1.DataType{Type: "INTEGER"}},
		{Name: "status", Type: flinkgatewayv1.DataType{Type: "VARCHAR"}},
	}
}

func testRow() types.StatementResultRow {
	return types.StatementResultRow{
		Operation: types.Insert,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{Type: types.Integer, Value: "1021"},
			types.AtomicStatementResultField{Type: types.Varchar, Value: "SHIPPED"},
		},
	}
}

func TestPrintQueryResult(t *testing.T) {
	t.Run("human table output with rows", func(t *testing.T) {
		c := newTestCommand(nil)
		cmd := newOutputCmd(t, "")
		result := &query.Result{
			Statement: flinkgatewayv1.SqlV1Statement{Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "COMPLETED"}},
			Columns:   testColumns(),
			Rows:      []types.StatementResultRow{testRow()},
		}

		out := captureStdout(t, func() {
			require.NoError(t, c.printQueryResult(cmd, "stmt", result, false, false))
		})
		require.Contains(t, out, "1021")
		require.Contains(t, out, "SHIPPED")
		require.NotContains(t, out, "Operation")
	})

	t.Run("human table output shows Operation column when requested", func(t *testing.T) {
		c := newTestCommand(nil)
		cmd := newOutputCmd(t, "")
		result := &query.Result{
			Statement: flinkgatewayv1.SqlV1Statement{Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "COMPLETED"}},
			Columns:   testColumns(),
			Rows:      []types.StatementResultRow{testRow()},
		}

		out := captureStdout(t, func() {
			require.NoError(t, c.printQueryResult(cmd, "stmt", result, true, false))
		})
		require.Contains(t, out, "Operation")
	})

	t.Run("human output with no rows prints a message to stderr, not stdout", func(t *testing.T) {
		c := newTestCommand(nil)
		cmd := newOutputCmd(t, "")
		result := &query.Result{
			Statement: flinkgatewayv1.SqlV1Statement{Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "COMPLETED"}},
		}

		out := captureStdout(t, func() {
			require.NoError(t, c.printQueryResult(cmd, "stmt", result, false, false))
		})
		require.Empty(t, out)
	})

	t.Run("json envelope carries schema, truncated and incomplete", func(t *testing.T) {
		c := newTestCommand(nil)
		cmd := newOutputCmd(t, "json")
		result := &query.Result{
			Statement:  flinkgatewayv1.SqlV1Statement{Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "RUNNING"}},
			Columns:    testColumns(),
			Rows:       []types.StatementResultRow{testRow()},
			Truncated:  true,
			Incomplete: true,
		}

		out := captureStdout(t, func() {
			require.NoError(t, c.printQueryResult(cmd, "stmt", result, false, false))
		})
		require.Contains(t, out, `"statement_name": "stmt"`)
		require.Contains(t, out, `"phase": "RUNNING"`)
		require.Contains(t, out, `"truncated": true`)
		require.Contains(t, out, `"incomplete": true`)
		require.Contains(t, out, `"id": 1021`)
	})

	t.Run("raw serialized output is a bare array with no envelope", func(t *testing.T) {
		c := newTestCommand(nil)
		cmd := newOutputCmd(t, "json")
		result := &query.Result{
			Statement: flinkgatewayv1.SqlV1Statement{Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "COMPLETED"}},
			Columns:   testColumns(),
			Rows:      []types.StatementResultRow{testRow()},
		}

		out := captureStdout(t, func() {
			require.NoError(t, c.printQueryResult(cmd, "stmt", result, false, true))
		})
		require.NotContains(t, out, "statement_name")
		require.Contains(t, out, `"id": 1021`)
	})

	t.Run("yaml serialized output", func(t *testing.T) {
		c := newTestCommand(nil)
		cmd := newOutputCmd(t, "yaml")
		result := &query.Result{
			Statement: flinkgatewayv1.SqlV1Statement{Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "COMPLETED"}},
			Columns:   testColumns(),
			Rows:      []types.StatementResultRow{testRow()},
		}

		out := captureStdout(t, func() {
			require.NoError(t, c.printQueryResult(cmd, "stmt", result, false, false))
		})
		require.Contains(t, out, "statement_name: stmt")
	})
}

// fakeJwtValidator lets tests control whether refreshGatewayToken thinks the
// current token is still valid without needing a real signed JWT.
type fakeJwtValidator struct {
	err error
}

func (f fakeJwtValidator) Validate(*cliconfig.Context) error {
	return f.err
}

func TestRefreshGatewayToken(t *testing.T) {
	t.Run("valid token is left alone", func(t *testing.T) {
		client := ccloudv2.NewFlinkGatewayClient("http://unused.invalid", "test", false, "still-valid")
		c := newTestCommand(newTestContext("http://unused.invalid", "still-valid"))

		refresh := c.refreshGatewayToken(client, fakeJwtValidator{err: nil})
		require.NoError(t, refresh())
		require.Equal(t, "still-valid", client.AuthToken)
	})

	t.Run("expired token is refreshed from the platform", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/api/access_tokens", r.URL.Path)
			require.Equal(t, "Bearer old-cloud-token", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"new-dataplane-token"}`))
		}))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient("http://unused.invalid", "test", false, "expired")
		c := newTestCommand(newTestContext(server.URL, "old-cloud-token"))

		refresh := c.refreshGatewayToken(client, fakeJwtValidator{err: errors.New("expired")})
		require.NoError(t, refresh())
		require.Equal(t, "new-dataplane-token", client.AuthToken)
	})

	t.Run("refresh failure surfaces the platform's error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error":"could not mint a dataplane token"}`))
		}))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient("http://unused.invalid", "test", false, "expired")
		c := newTestCommand(newTestContext(server.URL, "old-cloud-token"))

		refresh := c.refreshGatewayToken(client, fakeJwtValidator{err: errors.New("expired")})
		require.ErrorContains(t, refresh(), "could not mint a dataplane token")
		require.Equal(t, "expired", client.AuthToken)
	})
}

func TestStopStatement(t *testing.T) {
	t.Run("successful stop", func(t *testing.T) {
		server := httptest.NewServer(testserver.NewFlinkGatewayRouter(t))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient(server.URL, "test", false, "token")
		c := newTestCommand(newTestContext(server.URL, "token"))

		out := captureStderr(t, func() {
			require.True(t, c.stopStatement(client, "env-1", "stmt"))
		})
		require.Contains(t, out, `Stopped statement "stmt"`)
	})

	t.Run("failed stop reports a warning and returns false", func(t *testing.T) {
		// A 4xx, not 5xx: the retryable HTTP client retries 5xx/429 responses, which
		// would blow past stopTimeout and hit the timeout branch instead of this one.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient(server.URL, "test", false, "token")
		c := newTestCommand(newTestContext(server.URL, "token"))

		out := captureStderr(t, func() {
			require.False(t, c.stopStatement(client, "env-1", "stmt"))
		})
		require.Contains(t, out, `could not stop statement "stmt"`)
	})
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func TestHandleQueryError(t *testing.T) {
	t.Run("unbounded error stops the statement and names it in the suggestion", func(t *testing.T) {
		server := httptest.NewServer(testserver.NewFlinkGatewayRouter(t))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient(server.URL, "test", false, "token")
		c := newTestCommand(newTestContext(server.URL, "token"))

		settled := false
		var out string
		var err error
		out = captureStderr(t, func() {
			err = c.handleQueryError(client, "env-1", "stmt", &query.UnboundedError{StatementName: "stmt"}, &settled)
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unbounded result")
		var withSuggestions errors.ErrorWithSuggestions
		require.ErrorAs(t, err, &withSuggestions)
		require.Contains(t, withSuggestions.GetSuggestionsMsg(), `Statement "stmt" was stopped.`)
		require.True(t, settled)
		require.Contains(t, out, `Stopped statement "stmt"`)
	})

	t.Run("unbounded error names the manual stop command when the stop attempt fails", func(t *testing.T) {
		// A 4xx, not 5xx: see the equivalent comment in TestStopStatement.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient(server.URL, "test", false, "token")
		c := newTestCommand(newTestContext(server.URL, "token"))

		settled := false
		err := c.handleQueryError(client, "env-1", "stmt", &query.UnboundedError{StatementName: "stmt"}, &settled)
		require.Error(t, err)
		var withSuggestions errors.ErrorWithSuggestions
		require.ErrorAs(t, err, &withSuggestions)
		require.Contains(t, withSuggestions.GetSuggestionsMsg(), "confluent flink statement stop stmt")
		require.True(t, settled)
	})

	t.Run("context canceled is reported as interrupted", func(t *testing.T) {
		server := httptest.NewServer(testserver.NewFlinkGatewayRouter(t))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient(server.URL, "test", false, "token")
		c := newTestCommand(newTestContext(server.URL, "token"))

		settled := false
		err := c.handleQueryError(client, "env-1", "stmt", context.Canceled, &settled)
		require.Error(t, err)
		require.Contains(t, err.Error(), "interrupted")
		require.True(t, settled)
	})

	t.Run("context deadline exceeded is reported as timed out", func(t *testing.T) {
		server := httptest.NewServer(testserver.NewFlinkGatewayRouter(t))
		defer server.Close()

		client := ccloudv2.NewFlinkGatewayClient(server.URL, "test", false, "token")
		c := newTestCommand(newTestContext(server.URL, "token"))

		settled := false
		err := c.handleQueryError(client, "env-1", "stmt", context.DeadlineExceeded, &settled)
		require.Error(t, err)
		require.Contains(t, err.Error(), "timed out")
		require.True(t, settled)
	})

	t.Run("a 404 results-fetch error suggests the page-retention window", func(t *testing.T) {
		c := newTestCommand(nil)
		settled := false
		err := c.handleQueryError(nil, "env-1", "stmt", &query.ResultsFetchError{Err: flinkerror.NewError("not found", "", http.StatusNotFound)}, &settled)
		require.Error(t, err)
		var withSuggestions errors.ErrorWithSuggestions
		require.ErrorAs(t, err, &withSuggestions)
		require.Contains(t, withSuggestions.GetSuggestionsMsg(), "only retains result pages for a limited window")
		require.False(t, settled)
	})

	t.Run("a non-404 results-fetch error gets the generic suggestion", func(t *testing.T) {
		c := newTestCommand(nil)
		settled := false
		err := c.handleQueryError(nil, "env-1", "stmt", &query.ResultsFetchError{Err: flinkerror.NewError("boom", "", http.StatusInternalServerError)}, &settled)
		require.Error(t, err)
		var withSuggestions errors.ErrorWithSuggestions
		require.ErrorAs(t, err, &withSuggestions)
		require.Contains(t, withSuggestions.GetSuggestionsMsg(), "confluent flink statement describe stmt")
		require.False(t, settled)
	})

	t.Run("any other error falls back to the generic suggestion", func(t *testing.T) {
		c := newTestCommand(nil)
		settled := false
		err := c.handleQueryError(nil, "env-1", "stmt", errors.New("some other failure"), &settled)
		require.Error(t, err)
		require.Contains(t, err.Error(), "some other failure")
		require.False(t, settled)
	})
}
