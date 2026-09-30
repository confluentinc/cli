// Package query runs a bounded ("snapshot") Flink SQL statement to completion and
// returns the whole result set. See README.md for the design.
package query

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/wait"
)

const (
	initialBackoff = 300 * time.Millisecond
	maxBackoff     = 2 * time.Second

	// A statement typically leaves PENDING in well under a second.
	awaitPollInterval = 500 * time.Millisecond

	// A placeholder: wait.Options requires a nonzero Timeout, but the real bound is
	// the caller's ctx deadline, which wait.Poll checks independently.
	unboundedPollTimeout = 24 * time.Hour
)

// Options configures a single run. Only Client, EnvironmentId and OrganizationId
// are required.
type Options struct {
	Client         ccloudv2.GatewayClientInterface
	EnvironmentId  string
	OrganizationId string

	// MaxRows caps how many rows are collected; 0 means no cap. Hitting it sets
	// Result.Truncated rather than silently dropping rows.
	MaxRows int

	// RequireBounded rejects a statement whose traits say it is unbounded, rather
	// than draining a stream that never ends.
	RequireBounded bool

	// RefreshToken refreshes the gateway token. It runs before every gateway call
	// (force=false, which is a no-op while the current token is still valid) and
	// again with force=true after an Unauthorized response, to mint a new token
	// even when the old one looked unexpired. nil means no refresh.
	RefreshToken func(force bool) error

	// sleep is swapped out in tests so they do not wait in real time.
	sleep func(context.Context, time.Duration) error

	// pollInterval overrides awaitPollInterval in tests. Zero means use the
	// default.
	pollInterval time.Duration
}

// gatewayCall runs a gateway call with token handling: a proactive refresh
// before the call (non-fatal — the retry below is the real safety net), and, on
// an Unauthorized response, a forced token refresh and one retry. A long drain
// can outlive the dataplane token's lifetime; without the retry the call that
// straddles expiry fails with a bare Unauthorized and the whole run dies.
//
// The forced refresh and retry are safe even in a goroutine wait.Call has
// abandoned: the gateway client serializes its own token, so this retry's token
// read can't race the deferred stop's token write. The ctx check below is only
// an optimization — once ctx is done the caller has moved on to cleanup and will
// discard this result, so there's no point minting a token and retrying.
func gatewayCall[T any](ctx context.Context, opts Options, call func(ccloudv2.GatewayClientInterface) (T, error)) (T, error) {
	proactiveRefresh(opts)
	res, err := call(opts.Client)
	if err == nil || opts.RefreshToken == nil || !isUnauthorized(err) {
		return res, err
	}
	// ctx is already done: the caller abandoned this goroutine and will discard
	// the result, so skip the forced mint and retry as useless work.
	if ctxErr := ctx.Err(); ctxErr != nil {
		return res, ctxErr
	}
	// The token was accepted earlier but the gateway now rejects it — it lapsed
	// mid-run. Force a fresh token and retry the call once. A forced refresh that
	// itself fails, or a retry the gateway still rejects, means the cloud login
	// has expired: surface it as an AuthError so callers stop retrying and prompt
	// a re-login instead of polling to the deadline.
	if refreshErr := opts.RefreshToken(true); refreshErr != nil {
		return res, &AuthError{Err: refreshErr}
	}
	res, err = call(opts.Client)
	if err != nil && isUnauthorized(err) {
		return res, &AuthError{Err: err}
	}
	return res, err
}

// proactiveRefresh runs the best-effort, non-forced token refresh both gateway
// helpers do before a call. It never aborts the call: a failure is logged and
// the existing token is used, since gatewayCall's 401 retry is the real safety
// net. A nil RefreshToken (tests, callers that don't manage a token) is a no-op.
func proactiveRefresh(opts Options) {
	if opts.RefreshToken != nil {
		if err := opts.RefreshToken(false); err != nil {
			log.CliLogger.Warnf("Failed to refresh Flink gateway token: %v", err)
		}
	}
}

// bestEffortCall is gatewayCall without the forced re-mint and one-shot retry:
// just a proactive refresh then the call. For a status re-read whose result is
// discarded on any error (refreshStatement), a 401 isn't worth a forced
// /api/access_tokens round-trip and a second call — the caller keeps the value
// it already had.
func bestEffortCall[T any](opts Options, call func(ccloudv2.GatewayClientInterface) (T, error)) (T, error) {
	proactiveRefresh(opts)
	return call(opts.Client)
}

// isUnauthorized reports whether err carries an HTTP 401 from the gateway.
func isUnauthorized(err error) bool {
	var coder flinkerror.Coder
	return errors.As(err, &coder) && coder.StatusCode() == http.StatusUnauthorized
}

// AuthError reports that the gateway rejected the token and a fresh one could
// not be obtained — either the forced refresh failed or the retry was still
// Unauthorized. It means the user's cloud login has expired and they must log
// in again. Unlike a transient gateway error it is terminal: polling cannot
// recover from it, so await treats it as fatal rather than retrying to the
// deadline.
type AuthError struct {
	Err error
}

func (e *AuthError) Error() string { return e.Err.Error() }

func (e *AuthError) Unwrap() error { return e.Err }

// isAuthError reports whether err is (or wraps) an AuthError.
func isAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// Result is the outcome of a completed run.
type Result struct {
	// Statement as the gateway last reported it, including status and traits.
	Statement flinkgatewayv1.SqlV1Statement
	// Columns is the result schema, in order.
	Columns []flinkgatewayv1.ColumnDetails
	// Rows is the raw changelog as delivered. Every row is an insert for a bounded
	// append-only snapshot; otherwise the caller decides how to materialize it.
	Rows []types.StatementResultRow
	// Truncated reports that MaxRows stopped the drain before the result set ended.
	Truncated bool
}

// Phase is the statement phase at the end of the run.
func (r *Result) Phase() types.PHASE {
	return types.PHASE(r.Statement.Status.GetPhase())
}

// UnboundedError is returned when RequireBounded is set and the gateway reports the
// statement produces an unbounded result.
type UnboundedError struct {
	StatementName string
}

func (e *UnboundedError) Error() string {
	return fmt.Sprintf(`statement "%s" produces an unbounded result and cannot be run as a snapshot query`, e.StatementName)
}

// ResultsFetchError distinguishes a failed page fetch from a failed status read.
type ResultsFetchError struct {
	Err error
}

func (e *ResultsFetchError) Error() string {
	return e.Err.Error()
}

func (e *ResultsFetchError) Unwrap() error {
	return e.Err
}

// Run waits for an already-submitted statement to start, then drains every result page.
func Run(ctx context.Context, opts Options, statementName string) (*Result, error) {
	if opts.sleep == nil {
		opts.sleep = sleepContext
	}
	if opts.pollInterval == 0 {
		opts.pollInterval = awaitPollInterval
	}

	statement, err := await(ctx, opts, statementName)
	if err != nil {
		return nil, err
	}

	result := &Result{Statement: statement}

	traits := statement.Status.GetTraits()
	statementTraits := types.StatementTraits{FlinkGatewayV1StatementTraits: &traits}
	if isBounded, known := statementTraits.GetIsBounded(); opts.RequireBounded && known && !isBounded {
		return result, &UnboundedError{StatementName: statementName}
	}

	schema := traits.GetSchema()
	result.Columns = schema.GetColumns()

	// DDL/INSERT INTO statements have no schema and nothing to poll; return early
	// rather than reporting an empty table.
	if len(result.Columns) == 0 {
		return result, nil
	}

	if err := drain(ctx, opts, statementName, schema, result); err != nil {
		return result, err
	}

	return result, nil
}

// await polls until the statement leaves PENDING, so its traits (schema, boundedness) are populated.
func await(ctx context.Context, opts Options, statementName string) (flinkgatewayv1.SqlV1Statement, error) {
	return wait.PollPhases(ctx, wait.PhaseOptions[flinkgatewayv1.SqlV1Statement]{
		Fetch: func() (flinkgatewayv1.SqlV1Statement, error) {
			return wait.Call(ctx, func() (flinkgatewayv1.SqlV1Statement, error) {
				return gatewayCall(ctx, opts, func(c ccloudv2.GatewayClientInterface) (flinkgatewayv1.SqlV1Statement, error) {
					return c.GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId)
				})
			})
		},
		Phase:         func(s flinkgatewayv1.SqlV1Statement) string { return s.Status.GetPhase() },
		PendingPhases: []string{string(types.PENDING)},
		// A lapsed cloud login can't be waited out — surface it now instead of
		// re-minting a token every poll until the deadline.
		IsFatalErr:   isAuthError,
		PollInterval: opts.pollInterval,
		Timeout:      unboundedPollTimeout,
	})
}

// drain pulls pages until the gateway omits the next-page token — the authoritative
// "no more rows" signal; checking phase instead caused a real false-positive.
func drain(ctx context.Context, opts Options, statementName string, schema flinkgatewayv1.SqlV1ResultSchema, result *Result) error {
	pageToken := ""
	backoff := initialBackoff

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		pageRows, nextPageToken, err := fetchPage(ctx, opts, statementName, schema, pageToken)
		if err != nil {
			return err
		}
		result.Rows = append(result.Rows, pageRows...)

		if opts.MaxRows > 0 && len(result.Rows) > opts.MaxRows {
			result.Rows = result.Rows[:opts.MaxRows]
			result.Truncated = true
			refreshStatement(ctx, opts, statementName, result)
			return nil
		}

		if nextPageToken == "" {
			refreshStatement(ctx, opts, statementName, result)
			return nil
		}
		pageToken = nextPageToken

		backoff, err = nextBackoff(ctx, opts, len(pageRows) > 0, backoff)
		if err != nil {
			return err
		}
	}
}

// fetchPage reads one results page, converts it, and returns its rows and the
// token for the next page (empty when the gateway reported no next page). A
// fetch or conversion failure is wrapped in ResultsFetchError so the caller can
// tell it apart from a status read; a malformed next-page URL is not.
func fetchPage(ctx context.Context, opts Options, statementName string, schema flinkgatewayv1.SqlV1ResultSchema, pageToken string) ([]types.StatementResultRow, string, error) {
	page, err := wait.Call(ctx, func() (flinkgatewayv1.SqlV1StatementResult, error) {
		return gatewayCall(ctx, opts, func(c ccloudv2.GatewayClientInterface) (flinkgatewayv1.SqlV1StatementResult, error) {
			return c.GetStatementResults(opts.EnvironmentId, statementName, opts.OrganizationId, pageToken)
		})
	})
	if err != nil {
		return nil, "", &ResultsFetchError{Err: err}
	}

	pageResults := page.GetResults()
	converted, err := results.ConvertToInternalResults(pageResults.GetData(), schema)
	if err != nil {
		return nil, "", &ResultsFetchError{Err: err}
	}

	metadata := page.GetMetadata()
	nextPageToken := ""
	if nextUrl := metadata.GetNext(); nextUrl != "" {
		nextPageToken, err = ccloudv2.ExtractPageToken(nextUrl)
		if err != nil {
			return nil, "", err
		}
	}

	return converted.GetRows(), nextPageToken, nil
}

// nextBackoff resets the backoff after a page that carried rows, and otherwise
// sleeps the current backoff (a token but no rows means "nothing new yet, keep
// polling") before doubling it up to maxBackoff, so idle waiting doesn't hammer
// the gateway.
func nextBackoff(ctx context.Context, opts Options, hadRows bool, backoff time.Duration) (time.Duration, error) {
	if hadRows {
		return initialBackoff, nil
	}
	if err := opts.sleep(ctx, backoff); err != nil {
		return 0, err
	}
	return min(backoff*2, maxBackoff), nil
}

// refreshStatement re-reads the statement so Phase() reflects where it actually
// landed. Best-effort: a failed refresh just keeps the prior value, so it uses
// bestEffortCall — a 401 here isn't worth a forced token mint for a value we
// discard on error.
func refreshStatement(ctx context.Context, opts Options, statementName string, result *Result) {
	statement, err := wait.Call(ctx, func() (flinkgatewayv1.SqlV1Statement, error) {
		return bestEffortCall(opts, func(c ccloudv2.GatewayClientInterface) (flinkgatewayv1.SqlV1Statement, error) {
			return c.GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId)
		})
	})
	if err == nil {
		result.Statement = statement
	}
}

// IsTerminal reports whether phase is one the statement cannot leave on its own.
// Exported so callers holding a Result can check this without re-deriving the list.
func IsTerminal(phase types.PHASE) bool {
	switch phase {
	case types.COMPLETED, types.FAILED, types.STOPPED, types.DELETING:
		return true
	}
	return false
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
