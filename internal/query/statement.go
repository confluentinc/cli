package query

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

const (
	// stopTimeout bounds how long we wait for a statement to stop after interrupt.
	stopTimeout = 5 * time.Second

	// createStatementGracePeriod bounds how long we wait for a cancelled CreateStatement to land, so it can still be stopped.
	createStatementGracePeriod = 10 * time.Second
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
func (c *command) stopStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()

	_, err := wait.Call(ctx, func() (struct{}, error) {
		c.authTokenMu.Lock()
		defer c.authTokenMu.Unlock()
		statement, err := client.GetStatement(environmentId, name, c.Context.LastOrgId)
		if err != nil {
			return struct{}{}, err
		}
		if statement.Spec == nil {
			return struct{}{}, fmt.Errorf(`statement "%s" has no spec`, name)
		}
		statement.Spec.Stopped = flinkgatewayv1.PtrBool(true)
		return struct{}{}, client.UpdateStatement(environmentId, name, c.Context.LastOrgId, statement)
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

// stopStatementAndReport is stopStatement plus a user-facing stderr line, for
// the deferred end-of-run cleanup: after an error or a truncated read the user
// needs to see whether the leftover statement was actually released.
func (c *command) stopStatementAndReport(client *ccloudv2.FlinkGatewayClient, environmentId, name string) bool {
	ok, err := c.stopStatement(client, environmentId, name)
	switch {
	case ok:
		output.ErrPrintf(false, "Successfully stopped statement \"%s\".\n", name)
	case goerrors.Is(err, errStopTimeout):
		output.ErrPrintf(false, "Warning: timed out trying to stop statement \"%s\".\n", name)
	default:
		output.ErrPrintf(false, "Warning: could not stop statement \"%s\": %v.\n", name, err)
	}
	return ok
}

// isInterrupted reports whether err is a Ctrl-C/SIGTERM cancellation or a
// --wait-timeout expiry — the two cases interruptedError knows how to report.
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
				"Try again, or raise the limit with the `--wait-timeout` flag.",
			)
		}
		fate := stopFate(name, stopped)
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`query timed out before statement "%s" finished`, name),
			fmt.Sprintf("%s Raise the limit with the `--wait-timeout` flag if this keeps happening.", fate),
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
func (c *command) handleQueryError(cmd *cobra.Command, client *ccloudv2.FlinkGatewayClient, environmentId, name string, err error, settled *bool) error {
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
// outliving the short-lived dataplane token dies on a 401 before --wait-timeout.
func (c *command) refreshGatewayToken(client *ccloudv2.FlinkGatewayClient, jwtValidator jwt.Validator) func() error {
	return func() error {
		c.authTokenMu.Lock()
		defer c.authTokenMu.Unlock()

		jwtCtx := &cliconfig.Context{State: &cliconfig.ContextState{AuthToken: client.AuthToken}}
		if jwtValidator.Validate(jwtCtx) == nil {
			return nil
		}

		dataplaneToken, err := auth.GetDataplaneToken(c.Context)
		if err != nil {
			return err
		}
		client.AuthToken = dataplaneToken
		return nil
	}
}
