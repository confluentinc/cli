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
)

const (
	// stopTimeout bounds how long we wait for a statement to stop after interrupt.
	stopTimeout = 5 * time.Second

	// createStatementGracePeriod bounds how long we wait for a cancelled CreateStatement to land, so it can still be stopped.
	createStatementGracePeriod = 10 * time.Second
)

// createStatement runs create; on cancellation it waits up to gracePeriod for it
// to land and calls cleanup instead of leaking an unstoppable statement. The
// returned bool is cleanup's result (false if it was never called), so the
// caller can report the outcome itself instead of cleanup announcing it too.
func createStatement(ctx context.Context, gracePeriod time.Duration, create func() (flinkgatewayv1.SqlV1Statement, error), cleanup func() bool) (bool, error) {
	done := make(chan error, 1)
	go func() {
		_, err := create()
		done <- err
	}()

	select {
	case err := <-done:
		return false, err
	case <-ctx.Done():
		stopped := false
		select {
		case err := <-done:
			if err == nil {
				stopped = cleanup()
			}
		case <-time.After(gracePeriod):
		}
		return stopped, ctx.Err()
	}
}

// stopStatement makes a best-effort, bounded attempt to stop an abandoned
// statement. The gateway rejects a spec.stopped-only body, so this reads the
// statement back before flipping it. announce controls stderr vs. debug-log
// output: routine end-of-run cleanup logs quietly, interrupts/errors print.
func (c *command) stopStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string, announce bool) bool {
	done := make(chan error, 1)
	go func() {
		c.authTokenMu.Lock()
		defer c.authTokenMu.Unlock()
		done <- func() error {
			statement, err := client.GetStatement(environmentId, name, c.Context.LastOrgId)
			if err != nil {
				return err
			}
			if statement.Spec == nil {
				return fmt.Errorf(`statement "%s" has no spec`, name)
			}
			statement.Spec.Stopped = flinkgatewayv1.PtrBool(true)
			return client.UpdateStatement(environmentId, name, c.Context.LastOrgId, statement)
		}()
	}()

	report := func(warning, debug string) {
		if announce {
			output.ErrPrintf(false, "%s\n", warning)
		} else {
			log.CliLogger.Debugf("%s", debug)
		}
	}

	select {
	case err := <-done:
		if err != nil {
			report(
				fmt.Sprintf(`Warning: could not stop statement "%s": %v.`, name, err),
				fmt.Sprintf(`could not stop statement "%s" after query completion: %v`, name, err),
			)
			return false
		}
		report(
			fmt.Sprintf(`Successfully stopped statement "%s".`, name),
			fmt.Sprintf(`stopped statement "%s" after query completion`, name),
		)
		return true
	case <-time.After(stopTimeout):
		report(
			fmt.Sprintf(`Warning: timed out trying to stop statement "%s".`, name),
			fmt.Sprintf(`timed out trying to stop statement "%s" after query completion`, name),
		)
		return false
	}
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
		// announce=false: the outcome is folded into this one error's suggestion
		// below instead of also being printed separately by stopStatement, which
		// would say "stopped" a second time right next to this same message.
		fate := stopFate(name, c.stopStatement(client, environmentId, name, false))
		*settled = true
		return errors.NewErrorWithSuggestions(
			err.Error(),
			fmt.Sprintf("Bound the query with a LIMIT clause or a time predicate and run `confluent flink query` again. %s", fate),
		)
	}

	if goerrors.Is(err, context.Canceled) || goerrors.Is(err, context.DeadlineExceeded) {
		// announce=false for the same reason as above: interruptedError below states
		// the stop outcome itself, so stopStatement shouldn't also print it — that
		// previously produced a confusing "Successfully stopped ..." line immediately
		// followed by an unrelated "Error: query interrupted ..." for the same event.
		stopped := c.stopStatement(client, environmentId, name, false)
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
