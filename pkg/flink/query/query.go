// Package query runs a bounded ("snapshot") Flink SQL statement to completion,
// synchronously, and returns the whole result set.
//
// It does not reuse the interactive shell's Store/ResultFetcher stack: that
// pipeline silently evicts rows past a 10,000-row cap, drops schema-mismatched
// rows, and treats a missing page token as "done" without checking phase — all
// silent data loss once a script reads stdout. This package's drain loop reports
// each condition instead.
package query

import (
	"context"
	"fmt"
	"time"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/wait"
)

const (
	initialBackoff = 300 * time.Millisecond
	maxBackoff     = 2 * time.Second

	// awaitPollInterval is short relative to compute-pool-style polling (seconds,
	// not minutes): a statement typically leaves PENDING well under a second
	// after the gateway schedules it onto a compute pool.
	awaitPollInterval = 500 * time.Millisecond

	// unboundedPollTimeout exists only because wait.Options requires a nonzero
	// Timeout. The real bound is the ctx deadline the caller already set from
	// --wait-timeout, which wait.Poll checks independently via ctx.Done().
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

	// RefreshToken runs before every gateway call, since the dataplane token is
	// short-lived. Failures are logged, not returned — the next gateway call
	// surfaces its own error if the token is truly bad. Nil means no refresh.
	RefreshToken func() error

	// sleep is swapped out in tests so they do not wait in real time.
	sleep func(context.Context, time.Duration) error

	// pollInterval overrides awaitPollInterval in tests. Zero means use the
	// default.
	pollInterval time.Duration
}

// authenticatedClient refreshes the token if configured, mirroring
// Store.authenticatedGatewayClient.
func (opts Options) authenticatedClient() ccloudv2.GatewayClientInterface {
	if opts.RefreshToken != nil {
		if err := opts.RefreshToken(); err != nil {
			log.CliLogger.Warnf("Failed to refresh Flink gateway token: %v", err)
		}
	}
	return opts.Client
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

// ResultsFetchError distinguishes a failed page fetch from a failed
// statement-status read, so the caller can give a more specific suggestion for
// the gateway's page-retention window.
type ResultsFetchError struct {
	Err error
}

func (e *ResultsFetchError) Error() string {
	return e.Err.Error()
}

func (e *ResultsFetchError) Unwrap() error {
	return e.Err
}

// Run waits for an already-submitted statement to start, then drains every result
// page. The statement must already exist — submitting it is the caller's job, so the
// caller keeps ownership of naming, properties and cleanup.
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

// await polls until the statement leaves PENDING, so that its traits — the result
// schema and the boundedness flag — are populated. Uses the same wait.PollPhases
// helper other CLI commands use to block on a resource reaching a terminal phase
// (e.g. flink compute-pool create --wait), rather than a bespoke retry loop.
func await(ctx context.Context, opts Options, statementName string) (flinkgatewayv1.SqlV1Statement, error) {
	return wait.PollPhases(ctx, wait.PhaseOptions[flinkgatewayv1.SqlV1Statement]{
		Fetch: func() (flinkgatewayv1.SqlV1Statement, error) {
			return wait.Call(ctx, func() (flinkgatewayv1.SqlV1Statement, error) {
				return opts.authenticatedClient().GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId)
			})
		},
		Phase:         func(s flinkgatewayv1.SqlV1Statement) string { return s.Status.GetPhase() },
		PendingPhases: []string{string(types.PENDING)},
		PollInterval:  opts.pollInterval,
		Timeout:       unboundedPollTimeout,
	})
}

// drain pulls result pages until the gateway stops handing out a next-page
// token. That absence is itself the authoritative "no more rows" signal: the
// gateway's protocol guarantee is that it never omits the token while more rows
// remain, so unlike an earlier version of this loop, no statement-phase check is
// needed to decide whether the read is complete — the token already says so.
// Checking phase instead of trusting the token is what caused a real false
// positive: a bounded, LIMIT-satisfied query delivering every requested row
// still reported "may be incomplete", because the statement's own phase can
// stay RUNNING long after (or even indefinitely after) the last row is served —
// nothing about "no more rows" implies the job has reached a terminal phase.
//
// page_token is a positional offset, not a real cursor, so nothing can advance
// past a token-less page — that's fine, since a token-less page is final. A page
// with a token but no rows means "nothing new yet, keep polling this cursor";
// back off between empty pages so an idle wait doesn't hammer the gateway.
func drain(ctx context.Context, opts Options, statementName string, schema flinkgatewayv1.SqlV1ResultSchema, result *Result) error {
	pageToken := ""
	backoff := initialBackoff

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		page, err := wait.Call(ctx, func() (flinkgatewayv1.SqlV1StatementResult, error) {
			return opts.authenticatedClient().GetStatementResults(opts.EnvironmentId, statementName, opts.OrganizationId, pageToken)
		})
		if err != nil {
			return &ResultsFetchError{Err: err}
		}

		pageResults := page.GetResults()
		converted, err := results.ConvertToInternalResults(pageResults.GetData(), schema)
		if err != nil {
			return err
		}
		pageRows := converted.GetRows()
		result.Rows = append(result.Rows, pageRows...)

		if opts.MaxRows > 0 && len(result.Rows) > opts.MaxRows {
			result.Rows = result.Rows[:opts.MaxRows]
			result.Truncated = true
			refreshStatement(ctx, opts, statementName, result)
			return nil
		}

		metadata := page.GetMetadata()
		var nextPageToken string
		if nextUrl := metadata.GetNext(); nextUrl != "" {
			nextPageToken, err = ccloudv2.ExtractPageToken(nextUrl)
			if err != nil {
				return err
			}
		}

		if nextPageToken == "" {
			refreshStatement(ctx, opts, statementName, result)
			return nil
		}
		pageToken = nextPageToken

		if len(pageRows) == 0 {
			if err := opts.sleep(ctx, backoff); err != nil {
				return err
			}
			backoff = min(backoff*2, maxBackoff)
		} else {
			backoff = initialBackoff
		}
	}
}

// refreshStatement re-reads the statement once draining is done, so
// Result.Statement/Phase() reflect where things actually landed (in particular
// FAILED) rather than the value from before draining started. Best-effort: a
// failed refresh leaves the prior statement in place rather than losing the rows
// already collected — the caller still has a usable Result either way.
func refreshStatement(ctx context.Context, opts Options, statementName string, result *Result) {
	statement, err := wait.Call(ctx, func() (flinkgatewayv1.SqlV1Statement, error) {
		return opts.authenticatedClient().GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId)
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
