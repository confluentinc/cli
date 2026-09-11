package query

import (
	"context"
	goerrors "errors"
	"fmt"
	"net/http"
	"time"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	cliconfig "github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/output"
)

const (
	// stopTimeout bounds how long we wait for a statement to stop after interrupt.
	stopTimeout = 5 * time.Second

	// createStatementGracePeriod bounds how long we wait for a cancelled CreateStatement to land, so it can still be stopped.
	createStatementGracePeriod = 10 * time.Second
)

// createStatement runs create; on cancellation it waits up to gracePeriod for it to
// land and calls cleanup instead of leaking an unstoppable statement.
func createStatement(ctx context.Context, gracePeriod time.Duration, create func() (flinkgatewayv1.SqlV1Statement, error), cleanup func()) error {
	done := make(chan error, 1)
	go func() {
		_, err := create()
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		select {
		case err := <-done:
			if err == nil {
				cleanup()
			}
		case <-time.After(gracePeriod):
		}
		return ctx.Err()
	}
}

// stopStatement makes a best-effort, bounded attempt to stop an abandoned
// statement and reports the outcome either way. The gateway rejects a
// spec.stopped-only body, so this reads the statement back before flipping it.
func (c *command) stopStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string) bool {
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

	select {
	case err := <-done:
		if err != nil {
			output.ErrPrintf(false, "Warning: could not stop statement \"%s\": %v. It may still be running.\n", name, err)
			return false
		}
		output.ErrPrintf(false, "Stopped statement \"%s\".\n", name)
		return true
	case <-time.After(stopTimeout):
		output.ErrPrintf(false, "Warning: timed out trying to stop statement \"%s\". It may still be running.\n", name)
		return false
	}
}

// handleQueryError turns a failed or interrupted run into a message naming the
// statement; settled marks whether this call already stopped it.
func (c *command) handleQueryError(client *ccloudv2.FlinkGatewayClient, environmentId, name string, err error, settled *bool) error {
	var unbounded *query.UnboundedError
	if goerrors.As(err, &unbounded) {
		// Only claim the statement was stopped when it actually was.
		fate := fmt.Sprintf("Statement \"%s\" was stopped.", name)
		if !c.stopStatement(client, environmentId, name) {
			fate = fmt.Sprintf("Stop it with `confluent flink statement stop %s`.", name)
		}
		*settled = true
		return errors.NewErrorWithSuggestions(
			err.Error(),
			fmt.Sprintf("Bound the query with a LIMIT clause or a time predicate and run `confluent flink query` again. %s", fate),
		)
	}

	if goerrors.Is(err, context.Canceled) || goerrors.Is(err, context.DeadlineExceeded) {
		c.stopStatement(client, environmentId, name)
		*settled = true
		reason := "interrupted"
		if goerrors.Is(err, context.DeadlineExceeded) {
			reason = "timed out"
		}
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`query %s before statement "%s" finished`, reason, name),
			fmt.Sprintf("Check the statement with `confluent flink statement describe %s`, or raise the limit with the `--wait-timeout` flag.", name),
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
				return errors.NewErrorWithSuggestions(
					resultsFetchErr.Error(),
					fmt.Sprintf("Statement \"%s\" no longer exists — it may have been deleted, or the name is mistyped. Check `confluent flink statement describe %s`.", name, name),
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
			fmt.Sprintf("Check the statement with `confluent flink statement describe %s`.", name),
		)
	}

	return errors.NewErrorWithSuggestions(
		err.Error(),
		fmt.Sprintf("Check the statement with `confluent flink statement describe %s`.", name),
	)
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
