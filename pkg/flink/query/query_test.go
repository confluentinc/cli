package query

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

const (
	testEnvironmentId  = "env-123456"
	testOrganizationId = "org-123456"
	testStatementName  = "test-statement"
)

// testOptions builds Options whose sleep is a no-op, so backoff never costs wall time.
func testOptions(client *mock.MockGatewayClientInterface) Options {
	return Options{
		Client:         client,
		EnvironmentId:  testEnvironmentId,
		OrganizationId: testOrganizationId,
		sleep:          func(context.Context, time.Duration) error { return nil },
	}
}

func schema(columnNames ...string) *flinkgatewayv1.SqlV1ResultSchema {
	columns := make([]flinkgatewayv1.ColumnDetails, len(columnNames))
	for i, name := range columnNames {
		columns[i] = flinkgatewayv1.ColumnDetails{
			Name: name,
			Type: flinkgatewayv1.DataType{Type: "VARCHAR", Nullable: false},
		}
	}
	return &flinkgatewayv1.SqlV1ResultSchema{Columns: &columns}
}

func statement(phase string, traits *flinkgatewayv1.SqlV1StatementTraits) flinkgatewayv1.SqlV1Statement {
	return flinkgatewayv1.SqlV1Statement{
		Name: flinkgatewayv1.PtrString(testStatementName),
		Spec: &flinkgatewayv1.SqlV1StatementSpec{Statement: flinkgatewayv1.PtrString("SELECT * FROM t;")},
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Phase:  phase,
			Traits: traits,
		},
	}
}

func boundedTraits(columnNames ...string) *flinkgatewayv1.SqlV1StatementTraits {
	return &flinkgatewayv1.SqlV1StatementTraits{
		SqlKind:      flinkgatewayv1.PtrString("SELECT"),
		IsBounded:    flinkgatewayv1.PtrBool(true),
		IsAppendOnly: flinkgatewayv1.PtrBool(true),
		Schema:       schema(columnNames...),
	}
}

// page builds a result page. nextPageToken of "" means the gateway reported no next
// page.
func page(nextPageToken string, rows ...[]any) flinkgatewayv1.SqlV1StatementResult {
	data := make([]any, len(rows))
	for i, row := range rows {
		data[i] = map[string]any{"op": float64(0), "row": row}
	}

	metadata := flinkgatewayv1.ResultListMeta{}
	if nextPageToken != "" {
		next := fmt.Sprintf("https://flink.us-east-1.aws.confluent.cloud/sql/v1/organizations/%s/environments/%s/statements/%s/results?page_token=%s",
			testOrganizationId, testEnvironmentId, testStatementName, nextPageToken)
		metadata.Next = &next
	}

	return flinkgatewayv1.SqlV1StatementResult{
		Metadata: metadata,
		Results:  &flinkgatewayv1.SqlV1StatementResultResults{Data: &data},
	}
}

func rowValues(t *testing.T, result *Result) [][]string {
	t.Helper()
	values := make([][]string, len(result.Rows))
	for i, row := range result.Rows {
		fields := make([]string, len(row.GetFields()))
		for j, field := range row.GetFields() {
			fields[j] = field.ToString()
		}
		values[i] = fields
	}
	return values
}

func TestRunDrainsASinglePage(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id", "status"))

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1", "SHIPPED"}, []any{"2", "PENDING"}), nil)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1", "SHIPPED"}, {"2", "PENDING"}}, rowValues(t, result))
	require.False(t, result.Truncated)
	require.False(t, result.Incomplete)
	require.Equal(t, types.COMPLETED, result.Phase())
}

func TestRunDrainsEveryPage(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))
	completed := statement("COMPLETED", boundedTraits("id"))

	// Leaves PENDING, produces, completes after the last page — phase read before
	// each page, not after.
	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("10", []any{"1"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "10").
			Return(page("20", []any{"2"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "20").
			Return(page("", []any{"3"}), nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}, {"2"}, {"3"}}, rowValues(t, result))
	require.False(t, result.Incomplete)
}

// Unlike the shell (missing token = done), a run must not silently succeed when
// the statement is still running and there's no token left to advance with.
func TestRunFlagsIncompleteWhenPagesStopBeforeStatementDoes(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
		// The re-check before conceding Incomplete: still RUNNING, so it stays Incomplete.
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.True(t, result.Incomplete)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
}

// The statement can reach a terminal phase during the GetStatementResults call
// itself (reproduced against a real gateway) — terminalBeforeFetch is stale by
// then, so drain must re-read phase before concluding the read was short.
func TestRunDoesNotFlagIncompleteWhenStatementCompletesDuringTheFinalFetch(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))
	completed := statement("COMPLETED", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
		// The statement finished during the GetStatementResults call above.
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.False(t, result.Incomplete)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
}

// An empty page with no token is the gateway saying "nothing yet". Re-requesting the
// same offset is safe, so the loop should keep waiting rather than declaring victory.
func TestRunRetriesEmptyPagesUntilStatementIsTerminal(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))
	completed := statement("COMPLETED", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").Return(page(""), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.False(t, result.Incomplete)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
}

// Phase must be read before fetching a page, not after — reading it after could
// miss a completion happening in the gap. Pinning the call order here makes a
// regression fail loudly instead of occasionally under-counting rows in prod.
func TestRunReadsStatementStateBeforeFetchingResults(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))
	completed := statement("COMPLETED", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.False(t, result.Incomplete)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
}

func TestRunWaitsForPendingStatement(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	pending := statement("PENDING", nil)
	completed := statement("COMPLETED", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(pending, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(pending, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
}

func TestRunTruncatesAtMaxRowsAndSaysSo(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("10", []any{"1"}, []any{"2"}, []any{"3"}), nil)

	options := testOptions(client)
	options.MaxRows = 2

	result, err := Run(context.Background(), options, testStatementName)
	require.NoError(t, err)
	require.True(t, result.Truncated)
	require.Equal(t, [][]string{{"1"}, {"2"}}, rowValues(t, result))
}

// Exactly MaxRows rows is not a truncation — nothing was dropped.
func TestRunDoesNotReportTruncationWhenRowCountEqualsMaxRows(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1"}, []any{"2"}), nil)

	options := testOptions(client)
	options.MaxRows = 2

	result, err := Run(context.Background(), options, testStatementName)
	require.NoError(t, err)
	require.False(t, result.Truncated)
	require.Len(t, result.Rows, 2)
}

func TestRunRejectsAnUnboundedStatement(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	traits := boundedTraits("id")
	traits.IsBounded = flinkgatewayv1.PtrBool(false)

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(statement("RUNNING", traits), nil)

	options := testOptions(client)
	options.RequireBounded = true

	_, err := Run(context.Background(), options, testStatementName)
	var unbounded *UnboundedError
	require.ErrorAs(t, err, &unbounded)
	require.Equal(t, testStatementName, unbounded.StatementName)
}

// Without RequireBounded the caller opted into draining whatever comes back.
func TestRunAllowsUnboundedStatementWhenNotRequired(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	traits := boundedTraits("id")
	traits.IsBounded = flinkgatewayv1.PtrBool(false)
	completed := statement("COMPLETED", traits)

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1"}), nil)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Len(t, result.Rows, 1)
}

// A row whose width doesn't match the schema must fail the run — unlike the shell,
// which drops such rows silently.
func TestRunFailsOnRowSchemaMismatch(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id", "status"))

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1"}), nil)

	_, err := Run(context.Background(), testOptions(client), testStatementName)
	require.ErrorContains(t, err, "does not match the provided schema")
}

// DDL and INSERT INTO have no result schema; there is nothing to poll.
func TestRunSkipsResultsForStatementWithoutSchema(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	traits := &flinkgatewayv1.SqlV1StatementTraits{SqlKind: flinkgatewayv1.PtrString("CREATE_TABLE")}

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(statement("COMPLETED", traits), nil)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Empty(t, result.Rows)
	require.Empty(t, result.Columns)
}

func TestRunStopsOnCancelledContext(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(statement("PENDING", nil), nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	options := testOptions(client)
	options.sleep = func(context.Context, time.Duration) error {
		cancel()
		return context.Canceled
	}

	_, err := Run(ctx, options, testStatementName)
	require.ErrorIs(t, err, context.Canceled)
}

// A page-fetch failure is wrapped so the caller can tell it apart from a failure
// reading the statement itself.
func TestRunWrapsResultsFetchErrors(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(flinkgatewayv1.SqlV1StatementResult{}, errors.New("page not found"))

	_, err := Run(context.Background(), testOptions(client), testStatementName)
	var resultsErr *ResultsFetchError
	require.ErrorAs(t, err, &resultsErr)
	require.ErrorContains(t, err, "page not found")
}

func TestRunPropagatesGatewayErrors(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(flinkgatewayv1.SqlV1Statement{}, errors.New("unauthorized"))

	_, err := Run(context.Background(), testOptions(client), testStatementName)
	require.ErrorContains(t, err, "unauthorized")
}
