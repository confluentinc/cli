package flink

import (
	"context"
	goerrors "errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	cliconfig "github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/wait"
)

// createStatementGracePeriod bounds how long we wait for a cancelled CreateStatement to land, so it can still be stopped.
const createStatementGracePeriod = 10 * time.Second

// stopTimeout bounds the two gateway calls of a stop (GetStatement + UpdateStatement).
// stopTokenMintTimeout separately bounds the pre-stop token mint so a slow
// /api/access_tokens POST can't eat into stopTimeout. Both are vars, not consts,
// so tests can shrink them. See finding #2 in the review.
var (
	stopTimeout          = 5 * time.Second
	stopTokenMintTimeout = 5 * time.Second
)

// errStopTimeout marks a stop attempt that ran past stopTimeout, so callers can
// tell a timeout apart from an outright failure when reporting the outcome.
var errStopTimeout = goerrors.New("timed out")

// createStatement runs create; on cancellation it waits up to gracePeriod for it
// to land and calls cleanup instead of leaking an unstoppable statement. The
// returned bool is cleanup's result (false if it was never called), so the
// caller can report the outcome itself instead of cleanup announcing it too.
//
// If create() itself fails (with or without ctx also being done around the same
// time), that real error is returned as-is, not ctx.Err() — a create failure
// means no statement was ever named on the server, so reporting it as an
// interruption would send the caller off naming and suggesting `stop`/`describe`
// on a statement that never existed.
func createStatement(ctx context.Context, gracePeriod time.Duration, create func() (flinkgatewayv1.SqlV1Statement, error), cleanup func() bool) (bool, error) {
	done := make(chan error, 1)
	go func() {
		_, err := create()
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return false, err
		}
		// create succeeded, but select is random: ctx may have been cancelled at
		// the same instant. Honor the interrupt now — stop the statement we just
		// made and report it as interrupted — rather than letting the caller
		// proceed into the query with an already-dead context.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return cleanup(), ctxErr
		}
		return false, nil
	case <-ctx.Done():
		select {
		case err := <-done:
			if err != nil {
				return false, err
			}
			return cleanup(), ctx.Err()
		case <-time.After(gracePeriod):
			return false, ctx.Err()
		}
	}
}

// stopStatement makes a best-effort, bounded attempt to stop an abandoned
// statement. The gateway rejects a spec.stopped-only body, so this reads the
// statement back before flipping it. It only debug-logs the outcome; callers
// that must surface it to the user either fold the returned bool into their own
// message via stopFate, or use stopStatementAndReport. The returned error is
// errStopTimeout on timeout, the underlying failure otherwise, and nil on success.
func (c *queryCommand) stopStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()

	refresh := c.refreshGatewayToken(client, jwt.NewValidator())

	// Finding #2: mint the dataplane token on its own budget, before and outside the
	// stop's. A long drain can leave the token lapsed, and minting does a real
	// /api/access_tokens POST (30s client timeout); keeping that inside stopTimeout
	// could starve the two gateway calls and time the whole stop out. Best-effort:
	// on failure or its own timeout we fall back to the existing token, and the 401
	// retry below is the real safety net. force=false — a short query still holds a
	// valid token and needn't hit the endpoint. No lock needed: client.SetAuthToken
	// and the reads inside GetStatement/UpdateStatement are serialized by the client
	// itself, so an abandoned drain goroutine still refreshing can't race this stop.
	mintCtx, mintCancel := context.WithTimeout(context.Background(), stopTokenMintTimeout)
	if _, tokenErr := wait.Call(mintCtx, func() (struct{}, error) {
		return struct{}{}, refresh(false)
	}); tokenErr != nil {
		log.CliLogger.Debugf(`could not refresh token before stopping statement "%s"; using the existing one: %v`, name, tokenErr)
	}
	mintCancel()

	_, err := wait.Call(ctx, func() (struct{}, error) {
		// Findings #3 & #4: force a fresh token and retry on a 401, mirroring the
		// drain path's gatewayCall. Without it a token that looks locally valid but
		// the gateway rejects (clock skew / server-side revocation), or one that
		// lapses between the GET and the UPDATE, leaves the statement un-stopped.
		var statement flinkgatewayv1.SqlV1Statement
		if err := callWithForcedRefreshOn401(ctx, refresh, func() error {
			var e error
			statement, e = client.GetStatement(environmentId, name, c.Context.LastOrgId)
			return e
		}); err != nil {
			return struct{}{}, err
		}
		if statement.Spec == nil {
			return struct{}{}, fmt.Errorf(`statement "%s" has no spec`, name)
		}
		statement.Spec.Stopped = flinkgatewayv1.PtrBool(true)
		return struct{}{}, callWithForcedRefreshOn401(ctx, refresh, func() error {
			return client.UpdateStatement(environmentId, name, c.Context.LastOrgId, statement)
		})
	})

	switch {
	case err == nil:
		log.CliLogger.Debugf(`stopped statement "%s" after query completion`, name)
		return true, nil
	case goerrors.Is(err, context.DeadlineExceeded):
		log.CliLogger.Debugf(`timed out trying to stop statement "%s" after query completion`, name)
		return false, errStopTimeout
	default:
		log.CliLogger.Debugf(`could not stop statement "%s" after query completion: %v`, name, err)
		return false, err
	}
}

// callWithForcedRefreshOn401 runs a stop-path gateway call and, on a 401, forces a
// fresh dataplane token and retries once — the stop path's equivalent of the drain
// path's query.gatewayCall. Non-401 errors pass through untouched, as does the case
// where ctx is already done (the stop budget is spent; don't start another call). A
// forced-refresh failure keeps the original 401 rather than masking it.
func callWithForcedRefreshOn401(ctx context.Context, refresh func(bool) error, call func() error) error {
	err := call()
	if err == nil || !isGatewayUnauthorized(err) {
		return err
	}
	if ctx.Err() != nil {
		return err
	}
	if refreshErr := refresh(true); refreshErr != nil {
		log.CliLogger.Debugf("could not force-refresh the dataplane token to retry a stop call: %v", refreshErr)
		return err
	}
	return call()
}

// isGatewayUnauthorized reports whether err carries an HTTP 401 from the gateway.
func isGatewayUnauthorized(err error) bool {
	var coder flinkerror.Coder
	return goerrors.As(err, &coder) && coder.StatusCode() == http.StatusUnauthorized
}

// stopStatementAndReport is stopStatement plus a user-facing stderr line, for
// the deferred end-of-run cleanup: after an error or a truncated read the user
// needs to see whether the leftover statement was actually released.
func (c *queryCommand) stopStatementAndReport(client *ccloudv2.FlinkGatewayClient, environmentId, name string) bool {
	ok, err := c.stopStatement(client, environmentId, name)
	if ok {
		output.ErrPrintf(false, "Successfully stopped statement \"%s\".\n", name)
	} else {
		reportStopFailure(name, err)
	}
	return ok
}

// reportStopFailure prints a stderr warning for a stop attempt that didn't
// succeed. It's used both by stopStatementAndReport and by the routine quiet
// cleanup, which stays silent on success but must still surface a failure: a
// statement left running keeps burning compute even when the query itself
// finished fine.
func reportStopFailure(name string, err error) {
	if goerrors.Is(err, errStopTimeout) {
		output.ErrPrintf(false, "Warning: timed out trying to stop statement \"%s\".\n", name)
	} else {
		output.ErrPrintf(false, "Warning: could not stop statement \"%s\": %v.\n", name, err)
	}
}

// isInterrupted reports whether err is a Ctrl-C/SIGTERM cancellation or a
// --timeout expiry — the two cases interruptedError knows how to report.
func isInterrupted(err error) bool {
	return goerrors.Is(err, context.Canceled) || goerrors.Is(err, context.DeadlineExceeded)
}

// interruptedError turns a context cancellation/timeout into an actionable
// message. name is empty (stopped ignored) when interrupted before a statement
// exists. A timeout always reports "Error:"; Ctrl-C never does, since the user
// asked for it — it's "Interrupted:" whether or not the stop was confirmed.
func interruptedError(cmd *cobra.Command, err error, name string, stopped bool) error {
	if goerrors.Is(err, context.DeadlineExceeded) {
		if name == "" {
			return errors.NewErrorWithSuggestions(
				"query timed out before it started",
				"Try again, or raise the limit with the `--timeout` flag.",
			)
		}
		fate := stopFate(name, stopped)
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`query timed out before statement "%s" finished`, name),
			fmt.Sprintf("%s Raise the limit with the `--timeout` flag if this keeps happening.", fate),
		)
	}

	if name == "" {
		cmd.SilenceErrors = true
		output.ErrPrintln(false, "Interrupted: no statement had been created yet.")
		return errors.NewErrorWithSuggestions("query interrupted before it started", "")
	}
	if stopped {
		cmd.SilenceErrors = true
		output.ErrPrintf(false, "Interrupted: statement \"%s\" was stopped.\n", name)
		return errors.NewErrorWithSuggestions(fmt.Sprintf(`query interrupted before statement "%s" finished`, name), "")
	}

	msg := fmt.Sprintf(`query interrupted before statement "%s" finished`, name)
	cmd.SilenceErrors = true
	output.ErrPrintf(false, "Interrupted: %s\n", msg)
	return errors.NewErrorWithSuggestions(msg, stopFate(name, false))
}

// stopFate reports what happened to a statement in a single suggestion phrase:
// confirmed stopped, or how to stop it manually when that couldn't be confirmed.
func stopFate(name string, stopped bool) string {
	if stopped {
		return fmt.Sprintf(`Statement "%s" was stopped.`, name)
	}
	return fmt.Sprintf("Stop it with `confluent flink statement stop %s`.", name)
}

// handleQueryError turns a failed or interrupted run into a message naming the
// statement; settled marks whether this call already stopped it.
func (c *queryCommand) handleQueryError(cmd *cobra.Command, client *ccloudv2.FlinkGatewayClient, environmentId, name string, err error, settled *bool) error {
	var unbounded *query.UnboundedError
	if goerrors.As(err, &unbounded) {
		// The outcome is folded into this one error's suggestion below via
		// stopFate, instead of being printed separately, which would say
		// "stopped" a second time right next to this same message.
		stopped, _ := c.stopStatement(client, environmentId, name)
		fate := stopFate(name, stopped)
		*settled = true
		return errors.NewErrorWithSuggestions(
			err.Error(),
			fmt.Sprintf("Bound the query with a LIMIT clause or a time predicate and run `confluent flink query` again. %s", fate),
		)
	}

	if isInterrupted(err) {
		// Quiet on purpose: interruptedError below states the stop outcome itself,
		// so stopStatement shouldn't also print it — that previously produced a
		// confusing "Successfully stopped ..." line immediately followed by an
		// unrelated "Error: query interrupted ..." for the same event.
		stopped, _ := c.stopStatement(client, environmentId, name)
		*settled = true
		return interruptedError(cmd, err, name, stopped)
	}

	// The gateway rejected the token and a forced refresh couldn't recover it —
	// the cloud login itself has expired. Point the user at re-login rather than
	// the generic "describe the statement" suggestion, which wouldn't help.
	// Checked before ResultsFetchError because a drain-path AuthError arrives
	// wrapped in one; goerrors.As unwraps to find it either way.
	var authErr *query.AuthError
	if goerrors.As(err, &authErr) {
		// Mark settled so the caller's deferred cleanup doesn't fire its own stop:
		// the login is expired, so that stop can't mint a token either and would
		// 401, printing a contradictory "could not stop statement ... Unauthorized"
		// right after this re-login suggestion (the same reason the 404 branch below
		// sets settled). Fold the leftover statement into the suggestion instead.
		*settled = true
		return errors.NewErrorWithSuggestions(
			authErr.Error(),
			fmt.Sprintf("Log in again with `confluent login`, then stop statement \"%s\" with `confluent flink statement stop %s` if it is still running.", name, name),
		)
	}

	// Every other error, including ResultsFetchError below, leaves settled false:
	// the caller's deferred cleanup makes the stop attempt these branches skip.
	var resultsFetchErr *query.ResultsFetchError
	if goerrors.As(err, &resultsFetchErr) {
		var coder flinkerror.Coder
		if goerrors.As(err, &coder) {
			switch coder.StatusCode() {
			case http.StatusNotFound:
				// A 404 means the statement is gone or mistyped; an expired result window is a separate 408 (below).
				// It's already gone, so mark settled: the caller's deferred cleanup would
				// otherwise fire its own stop attempt and print a second, contradictory
				// 404 warning right after this message says the statement no longer exists.
				*settled = true
				return errors.NewErrorWithSuggestions(
					resultsFetchErr.Error(),
					fmt.Sprintf("Statement \"%s\" no longer exists — it may have been deleted, or the name is mistyped. Check %s.", name, describeCmd(name)),
				)
			case http.StatusRequestTimeout:
				// Results are retained for one hour; past that the gateway returns 408 with its own message.
				return errors.NewErrorWithSuggestions(
					resultsFetchErr.Error(),
					"Re-run the query — the result window for this statement has closed.",
				)
			}
		}
		return errors.NewErrorWithSuggestions(
			resultsFetchErr.Error(),
			fmt.Sprintf("Check the statement with %s.", describeCmd(name)),
		)
	}

	return errors.NewErrorWithSuggestions(
		err.Error(),
		fmt.Sprintf("Check the statement with %s.", describeCmd(name)),
	)
}

// describeCmd returns the backtick-quoted `confluent flink statement describe
// <name>` command text shared by several suggestion messages above.
func describeCmd(name string) string {
	return fmt.Sprintf("`confluent flink statement describe %s`", name)
}

// refreshGatewayToken mirrors the shell's pre-call check: without it, a query
// outliving the short-lived dataplane token dies on a 401 before --timeout.
// When force is set the not-yet-expired shortcut is skipped and a new token is
// always minted — used to recover after the gateway rejects a token that still
// looked valid locally.
func (c *queryCommand) refreshGatewayToken(client *ccloudv2.FlinkGatewayClient, jwtValidator jwt.Validator) func(bool) error {
	return func(force bool) error {
		return c.mintDataplaneToken(client, jwtValidator, force)
	}
}

// mintDataplaneToken swaps the client's auth token for a freshly-minted
// dataplane token. Unless force is set it first checks the current token, and
// skips the /api/access_tokens round-trip while it is still valid. A mint
// failure leaves the token untouched and is returned; the caller decides
// whether that is fatal (the query path) or best-effort (the stop path, which
// falls back to the existing token). The client serializes the token's read and
// write internally, so callers need no lock of their own.
func (c *queryCommand) mintDataplaneToken(client *ccloudv2.FlinkGatewayClient, jwtValidator jwt.Validator, force bool) error {
	if !force {
		jwtCtx := &cliconfig.Context{State: &cliconfig.ContextState{AuthToken: client.GetAuthToken()}}
		if jwtValidator.Validate(jwtCtx) == nil {
			return nil
		}
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.Context)
	if err != nil {
		return err
	}
	client.SetAuthToken(dataplaneToken)
	return nil
}
