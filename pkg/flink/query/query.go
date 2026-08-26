// Package query runs a bounded ("snapshot") Flink SQL statement to completion and
// returns the whole result set, synchronously, from the client.
//
// The Flink gateway has no synchronous execute endpoint: a statement is submitted,
// polled until it leaves PENDING, and then its result pages are pulled one at a
// time. This package performs that handshake behind a single call so a
// non-interactive command can behave like an ordinary database query.
//
// It deliberately does not reuse the interactive shell's Store/ResultFetcher stack.
// That pipeline is built for a scrolling viewer and degrades gracefully in three
// ways that are invisible to the user — it evicts rows past a 10,000-row cap, it
// drops rows whose width does not match the schema without reporting it, and it
// treats a missing page token as "done" without checking the statement phase. Every
// one of those turns into silent data loss once a script is consuming stdout, so the
// drain loop here is separate and reports each of those conditions instead.
package query

import (
	"context"
	"fmt"
	"net/url"
	"time"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
)

const (
	initialBackoff = 300 * time.Millisecond
	maxBackoff     = 2 * time.Second
)

// Options configures a single run. Only Client, EnvironmentId and OrganizationId
// are required.
type Options struct {
	Client         ccloudv2.GatewayClientInterface
	EnvironmentId  string
	OrganizationId string

	// MaxRows caps how many rows are collected. Zero means no cap. When the cap is
	// hit the run stops early and Result.Truncated is set — rows are never dropped
	// without saying so.
	MaxRows int

	// RequireBounded rejects a statement whose traits say it is unbounded, rather
	// than draining a stream that will never end.
	RequireBounded bool

	// RefreshToken is called before every gateway request. The dataplane token behind
	// Client is short-lived, and unlike the interactive shell — which wraps every call
	// in synchronizedTokenRefresh — a raw Client here would let a query that outlives
	// that token die on a 401. A refresh failure is logged and swallowed rather than
	// aborting the run: the next gateway call surfaces its own error if the token
	// really is bad, which is a clearer failure than one raised from inside a refresh
	// helper. Nil means there is nothing to refresh.
	RefreshToken func() error

	// sleep is swapped out in tests so they do not wait in real time.
	sleep func(context.Context, time.Duration) error
}

// authenticatedClient refreshes the gateway token, if a refresher was configured,
// before returning the client to call. Mirrors Store.authenticatedGatewayClient.
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
	// Rows is the raw changelog as delivered by the gateway. For a bounded
	// append-only snapshot every row is an insert; for anything else the caller has
	// to decide whether to materialize it.
	Rows []types.StatementResultRow
	// Truncated reports that MaxRows stopped the drain before the result set ended.
	Truncated bool
	// Incomplete reports that the gateway stopped handing out page tokens while the
	// statement was still running, so rows may be missing. See drain.
	Incomplete bool
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

// ResultsFetchError wraps a failure to fetch a result page, as distinct from a
// failure to read the statement's own status. The two calls fail for different
// reasons — a missing statement versus a page that no longer resolves — so keeping
// them distinguishable lets the caller give a more specific suggestion, in
// particular for the gateway's limited result-page retention window.
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

	// A statement that produces no result set at all (DDL, INSERT INTO) has no schema
	// and no results endpoint worth polling. Returning here keeps the caller from
	// reporting an empty table for a statement that legitimately has no rows.
	if len(result.Columns) == 0 {
		return result, nil
	}

	if err := drain(ctx, opts, statementName, schema, result); err != nil {
		return result, err
	}

	return result, nil
}

// await polls until the statement leaves PENDING, so that its traits — the result
// schema and the boundedness flag — are populated.
func await(ctx context.Context, opts Options, statementName string) (flinkgatewayv1.SqlV1Statement, error) {
	backoff := initialBackoff
	for {
		if err := ctx.Err(); err != nil {
			return flinkgatewayv1.SqlV1Statement{}, err
		}

		statement, err := opts.authenticatedClient().GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId)
		if err != nil {
			return flinkgatewayv1.SqlV1Statement{}, err
		}

		if types.PHASE(statement.Status.GetPhase()) != types.PENDING {
			return statement, nil
		}

		if err := opts.sleep(ctx, backoff); err != nil {
			return flinkgatewayv1.SqlV1Statement{}, err
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

// drain pulls result pages until the gateway reports no next page and the statement
// has reached a terminal phase.
//
// The statement's phase is read before each page, not after. Finishing was previously
// inferred from two facts read at two different moments — the page (with no next
// token) and then the phase — in that order. A statement that finished in the gap
// between them made a possibly-short page look complete: the phase read afterward
// could reflect a completion that happened after the page was already fetched, even
// though more rows might have existed in between. Reading the phase first means that
// when it is already terminal, nothing more can be added before the page is fetched,
// so an empty next token on that page is genuinely final. The cost is one extra status
// call per page.
//
// The gateway can return a page with no next token while the statement is still
// RUNNING, and the page token is a positional offset into the collect sink rather than
// an opaque cursor — so there is no token that advances past a page that did not
// supply one. When that happens on a page that carried rows and the statement was not
// already terminal, the result set may be short and Incomplete says so. When it
// happens on an empty page, re-requesting the same offset is harmless, so the loop
// backs off and retries.
func drain(ctx context.Context, opts Options, statementName string, schema flinkgatewayv1.SqlV1ResultSchema, result *Result) error {
	pageToken := ""
	backoff := initialBackoff

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		statement, err := opts.authenticatedClient().GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId)
		if err != nil {
			return err
		}
		result.Statement = statement
		terminalBeforeFetch := IsTerminal(types.PHASE(statement.Status.GetPhase()))

		page, err := opts.authenticatedClient().GetStatementResults(opts.EnvironmentId, statementName, opts.OrganizationId, pageToken)
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
			return nil
		}

		metadata := page.GetMetadata()
		nextPageToken, err := extractPageToken(metadata.GetNext())
		if err != nil {
			return err
		}

		if nextPageToken != "" {
			pageToken = nextPageToken
			backoff = initialBackoff
			continue
		}

		if terminalBeforeFetch {
			return nil
		}

		if len(pageRows) > 0 {
			// terminalBeforeFetch only reflects the phase read before the results
			// call above — the statement can reach a terminal phase during that
			// call, which would make this a false Incomplete on a query that
			// actually finished. Re-read the phase once before conceding; a
			// GetStatement failure here falls back to Incomplete rather than
			// losing the rows already collected.
			if statement, err := opts.authenticatedClient().GetStatement(opts.EnvironmentId, statementName, opts.OrganizationId); err == nil {
				result.Statement = statement
				if IsTerminal(types.PHASE(statement.Status.GetPhase())) {
					return nil
				}
			}
			result.Incomplete = true
			return nil
		}

		if err := opts.sleep(ctx, backoff); err != nil {
			return err
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

// IsTerminal reports whether phase is one the statement cannot leave on its own.
// Exported so callers that hold a Result after Run returns — success or error — can
// tell whether the statement still needs to be stopped without re-deriving this list.
func IsTerminal(phase types.PHASE) bool {
	switch phase {
	case types.COMPLETED, types.FAILED, types.STOPPED, types.DELETING:
		return true
	}
	return false
}

// extractPageToken pulls the page_token query parameter out of the absolute URL the
// gateway returns in metadata.next.
func extractPageToken(nextUrl string) (string, error) {
	if nextUrl == "" {
		return "", nil
	}

	parsed, err := url.Parse(nextUrl)
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", err
	}

	return params.Get("page_token"), nil
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
