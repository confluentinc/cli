package query

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
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
		pollInterval:   time.Millisecond,
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

	// One call for await (leaves PENDING immediately), one for drain's
	// end-of-loop refreshStatement — not a per-page phase check.
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).Times(2)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1", "SHIPPED"}, []any{"2", "PENDING"}), nil)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1", "SHIPPED"}, {"2", "PENDING"}}, rowValues(t, result))
	require.False(t, result.Truncated)
	require.Equal(t, types.COMPLETED, result.Phase())
}

func TestRunForcesTokenRefreshAndRetriesOnUnauthorized(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))
	unauthorized := flinkerror.NewError("Unauthorized", "", http.StatusUnauthorized)

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		// The dataplane token lapses mid-drain: the first page fetch is rejected...
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(flinkgatewayv1.SqlV1StatementResult{}, unauthorized),
		// ...and after a forced refresh the retry succeeds.
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
	)

	forced := 0
	opts := testOptions(client)
	opts.RefreshToken = func(force bool) error {
		if force {
			forced++
		}
		return nil
	}

	result, err := Run(context.Background(), opts, testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
	require.Equal(t, 1, forced, "expected exactly one forced refresh after the 401")
}

func TestRunSurfacesRefreshFailureWhenTokenIsRejected(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))
	unauthorized := flinkerror.NewError("Unauthorized", "", http.StatusUnauthorized)

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(flinkgatewayv1.SqlV1StatementResult{}, unauthorized),
	)

	opts := testOptions(client)
	opts.RefreshToken = func(force bool) error {
		if force {
			return errors.New("re-login required")
		}
		return nil
	}

	_, err := Run(context.Background(), opts, testStatementName)
	require.ErrorContains(t, err, "re-login required")
}

// Once ctx is done, gatewayCall's goroutine is abandoned by wait.Call but keeps
// running the shared, context-less gateway client. The forced refresh + retry
// are race-safe regardless (the client serializes its token), but they're wasted
// work in an abandoned goroutine whose result is discarded, so gatewayCall skips
// them. Tested directly because the leak is a background goroutine the
// higher-level Run test can't observe deterministically.
func TestGatewayCallSkipsForcedRefreshAndRetryOnceContextIsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	forced := 0
	calls := 0
	opts := Options{
		RefreshToken: func(force bool) error {
			if force {
				forced++
			}
			return nil
		},
	}
	unauthorized := flinkerror.NewError("Unauthorized", "", http.StatusUnauthorized)

	_, err := gatewayCall(ctx, opts, func(ccloudv2.GatewayClientInterface) (int, error) {
		calls++
		return 0, unauthorized
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 0, forced, "must not force a token refresh in an abandoned goroutine")
	require.Equal(t, 1, calls, "must not retry the call in an abandoned goroutine")
}

// A forced refresh whose retry the gateway still rejects is an expired cloud
// login, not a transient blip: gatewayCall must surface it as an AuthError so
// the caller prompts a re-login instead of retrying.
func TestGatewayCallReturnsAuthErrorWhenRetryStaysUnauthorized(t *testing.T) {
	forced := 0
	opts := Options{
		RefreshToken: func(force bool) error {
			if force {
				forced++
			}
			return nil
		},
	}
	unauthorized := flinkerror.NewError("Unauthorized", "", http.StatusUnauthorized)

	_, err := gatewayCall(context.Background(), opts, func(ccloudv2.GatewayClientInterface) (int, error) {
		return 0, unauthorized
	})

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	require.Equal(t, 1, forced, "expected exactly one forced refresh before giving up")
}

// A forced-refresh failure while the statement is still PENDING must surface
// immediately as an AuthError, not be swallowed as a transient fetch error that
// polls (and re-mints a token) every interval until --timeout. The single
// GetStatement expectation is the assertion: a retrying poll would call it again.
func TestRunSurfacesAuthErrorWhileAwaitingInsteadOfPolling(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	unauthorized := flinkerror.NewError("Unauthorized", "", http.StatusUnauthorized)
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(statement("PENDING", nil), unauthorized)

	forced := 0
	opts := testOptions(client)
	opts.RefreshToken = func(force bool) error {
		if force {
			forced++
			return errors.New("re-login required")
		}
		return nil
	}

	// Generous deadline: the test proves the run returns because the error is
	// fatal, not because the context expired.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	_, err := Run(ctx, opts, testStatementName)
	require.Less(t, time.Since(start), time.Second, "an expired login must abort the run, not wait out --timeout")

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	require.ErrorContains(t, err, "re-login required")
	require.Equal(t, 1, forced, "expected a single forced refresh, not one per poll")
}

// refreshStatement is best-effort: a 401 on the final status re-read must not
// trigger a forced token mint, since its result is discarded on error. The run
// still succeeds with the statement await already saw.
func TestRunDoesNotForceRefreshOnBestEffortStatementReread(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))
	unauthorized := flinkerror.NewError("Unauthorized", "", http.StatusUnauthorized)

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
		// The final refreshStatement re-read is rejected; best-effort, so it's dropped
		// without a forced mint or a retry.
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
			Return(flinkgatewayv1.SqlV1Statement{}, unauthorized),
	)

	forced := 0
	opts := testOptions(client)
	opts.RefreshToken = func(force bool) error {
		if force {
			forced++
		}
		return nil
	}

	result, err := Run(context.Background(), opts, testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
	require.Equal(t, types.COMPLETED, result.Phase(), "keeps the statement await saw when the re-read fails")
	require.Equal(t, 0, forced, "a best-effort re-read must not force a token mint on a 401")
}

func TestRunDrainsEveryPage(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))
	completed := statement("COMPLETED", boundedTraits("id"))

	// Pages are followed purely by token; the statement's phase is only read once
	// up front (await) and once at the end (refreshStatement), not per page.
	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("10", []any{"1"}), nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "10").
			Return(page("20", []any{"2"}), nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "20").
			Return(page("", []any{"3"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}, {"2"}, {"3"}}, rowValues(t, result))
	require.Equal(t, types.COMPLETED, result.Phase())
}

// Regression test for a real false positive found running this against staging:
// a bounded, LIMIT-satisfied query delivered every one of the requested rows,
// but the underlying job's phase stayed RUNNING — it never transitions to
// COMPLETED just because the row stream ended. A token-less page must be
// treated as done regardless of phase; there is no "recheck and maybe concede
// incomplete" step anymore, because the token itself is the authoritative
// completion signal.
func TestRunTreatsATokenLessPageAsDoneRegardlessOfPhase(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
	require.Equal(t, types.RUNNING, result.Phase())
}

// A page with a token but zero rows means "nothing new yet" — keep polling that
// cursor. Only a token-less page means done; an empty page that still hands out
// a token is not the same thing.
func TestRunRetriesEmptyPagesThatStillCarryAToken(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	running := statement("RUNNING", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").Return(page("5"), nil),
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "5").
			Return(page("", []any{"1"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(running, nil),
	)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
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
		client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
			Return(page("", []any{"1"}), nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil),
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

	// Only one call (await): the conversion error short-circuits drain before it
	// ever reaches the token check that would trigger refreshStatement.
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil)
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

// await no longer sleeps through Options.sleep (see wait.PollPhases); a
// cancellation arriving mid-poll must still stop the run. Simulated here by
// cancelling as a side effect of the GetStatement call itself, since
// wait.Poll only checks ctx.Done() between fetches, not sleep.
func TestRunStopsOnCancelledContext(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))

	ctx, cancel := context.WithCancel(context.Background())
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		DoAndReturn(func(string, string, string) (flinkgatewayv1.SqlV1Statement, error) {
			cancel()
			return statement("PENDING", nil), nil
		}).AnyTimes()

	_, err := Run(ctx, testOptions(client), testStatementName)
	require.ErrorIs(t, err, context.Canceled)
}

// The realistic --timeout case: the statement just never leaves PENDING
// in time, with no error at all. wait.Poll's own deadline never fires here
// (it's the 24h placeholder), so this exercises the ctx.Done() branch on a
// context whose deadline elapses naturally rather than one cancelled by a test
// side effect.
func TestRunTimesOutWhileStatementStaysPending(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(statement("PENDING", nil), nil).AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := Run(ctx, testOptions(client), testStatementName)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// Options.pollInterval is a private field defaulted in Run(), same as sleep;
// this exercises that default directly by constructing Options without going
// through testOptions. A statement that is already terminal on the first
// GetStatement call never reaches wait.Poll's ticker, so this also confirms
// the default doesn't itself break the zero-poll case (wait.Poll rejects a
// non-positive PollInterval outright).
func TestRunDefaultsPollInterval(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(statement("COMPLETED", nil), nil)

	_, err := Run(context.Background(), Options{
		Client:         client,
		EnvironmentId:  testEnvironmentId,
		OrganizationId: testOrganizationId,
	}, testStatementName)
	require.NoError(t, err)
}

// GatewayClientInterface takes no context and ignores any deadline the caller
// set (ccloudv2.FlinkGatewayClient builds every request from
// context.Background()). Confirmed against real staging: a GetStatementResults
// call once hung for 49 minutes despite a 2-minute --timeout. callWithContext
// is what makes Run return once ctx fires regardless — this pins that down by
// blocking GetStatementResults forever and asserting Run still returns promptly.
func TestRunReturnsPromptlyWhenResultsCallHangsPastDeadline(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil).AnyTimes()
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		DoAndReturn(func(string, string, string, string) (flinkgatewayv1.SqlV1StatementResult, error) {
			select {} // never returns; the shared client has no way to cancel this
		})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := Run(ctx, testOptions(client), testStatementName)
	elapsed := time.Since(start)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, elapsed, time.Second, "Run must not wait for the hung call once ctx fires")
}

// Same as above, for the GetStatement call await() makes.
func TestRunReturnsPromptlyWhenAwaitCallHangsPastDeadline(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		DoAndReturn(func(string, string, string) (flinkgatewayv1.SqlV1Statement, error) {
			select {} // never returns
		}).AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := Run(ctx, testOptions(client), testStatementName)
	elapsed := time.Since(start)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, elapsed, time.Second, "Run must not wait for the hung call once ctx fires")
}

// A page-fetch failure is wrapped so the caller can tell it apart from a failure
// reading the statement itself.
func TestRunWrapsResultsFetchErrors(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))

	// Only one call (await): a failed fetch returns immediately, before drain
	// ever reaches the token check that would trigger refreshStatement.
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(flinkgatewayv1.SqlV1StatementResult{}, errors.New("page not found"))

	_, err := Run(context.Background(), testOptions(client), testStatementName)
	var resultsErr *ResultsFetchError
	require.ErrorAs(t, err, &resultsErr)
	require.ErrorContains(t, err, "page not found")
}

// A page fetch can succeed but its rows still fail to convert (e.g. an
// unrecognized column type); that failure must be wrapped the same way as a
// network fetch failure, so handleQueryError's ResultsFetchError-specific
// suggestions (404/408) still apply instead of silently degrading to the
// generic fallback.
func TestRunWrapsConversionErrorsTheSameAsResultsFetchErrors(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	badSchema := &flinkgatewayv1.SqlV1ResultSchema{Columns: &[]flinkgatewayv1.ColumnDetails{
		{Name: "id", Type: flinkgatewayv1.DataType{Type: "NOT_A_REAL_TYPE", Nullable: false}},
	}}
	completed := flinkgatewayv1.SqlV1Statement{
		Name: flinkgatewayv1.PtrString(testStatementName),
		Spec: &flinkgatewayv1.SqlV1StatementSpec{Statement: flinkgatewayv1.PtrString("SELECT * FROM t;")},
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Phase: "COMPLETED",
			Traits: &flinkgatewayv1.SqlV1StatementTraits{
				SqlKind:      flinkgatewayv1.PtrString("SELECT"),
				IsBounded:    flinkgatewayv1.PtrBool(true),
				IsAppendOnly: flinkgatewayv1.PtrBool(true),
				Schema:       badSchema,
			},
		},
	}

	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).Return(completed, nil)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1"}), nil)

	_, err := Run(context.Background(), testOptions(client), testStatementName)
	var resultsErr *ResultsFetchError
	require.ErrorAs(t, err, &resultsErr)
	require.ErrorContains(t, err, "unsupported result field type")
}

// A GetStatement error while awaiting PENDING is transient by wait.Poll's design
// (matching every other CLI command built on it): it retries rather than
// aborting, so a persistent error surfaces once the caller's own context
// deadline fires, as ctx.Err() rather than the underlying error text. That
// deadline is what --timeout controls, and handleQueryError already turns
// it into a "query timed out" message.
func TestRunSurfacesContextDeadlineOnPersistentAwaitErrors(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
		Return(flinkgatewayv1.SqlV1Statement{}, errors.New("unauthorized")).AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := Run(ctx, testOptions(client), testStatementName)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// A transient GetStatement error while awaiting PENDING must not abort the
// run: wait.Poll retries until the statement is fetched successfully.
func TestRunRecoversFromTransientAwaitError(t *testing.T) {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	completed := statement("COMPLETED", boundedTraits("id"))

	gomock.InOrder(
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
			Return(flinkgatewayv1.SqlV1Statement{}, errors.New("temporary blip")),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
			Return(completed, nil),
		client.EXPECT().GetStatement(testEnvironmentId, testStatementName, testOrganizationId).
			Return(completed, nil),
	)
	client.EXPECT().GetStatementResults(testEnvironmentId, testStatementName, testOrganizationId, "").
		Return(page("", []any{"1"}), nil)

	result, err := Run(context.Background(), testOptions(client), testStatementName)
	require.NoError(t, err)
	require.Equal(t, [][]string{{"1"}}, rowValues(t, result))
}
