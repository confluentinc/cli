package wait

import (
	"context"
	"errors"
	"time"
)

// Call races fn against ctx, for an API like FlinkGatewayClient's that ignores context
// entirely. It can't abort fn once started, so its goroutine leaks until fn returns.
func Call[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	type result struct {
		val T
		err error
	}
	ch := make(chan result, 1)
	go func() {
		val, err := fn()
		ch <- result{val, err}
	}()

	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// Options describes a single Poll invocation. T is the polled resource type.
//
// Delay (if non-zero) sleeps before the first Fetch. Use it to give the
// resource a moment to materialize on the server after a POST; mirrors
// retry.StateChangeConf.Delay in terraform-provider-confluent.
//
// PollInterval is the gap between successive Fetch calls after the first.
//
// Initial (optional) seeds the last-known value Poll returns with an error: a
// caller that already holds the object (a create response) passes it so a
// timeout that fires before the first Fetch still reports a real phase rather
// than the zero value.
type Options[T any] struct {
	Fetch        func() (T, error)
	IsTerminal   func(T) bool
	IsFailed     func(T) bool
	Initial      T
	Delay        time.Duration
	PollInterval time.Duration
	Timeout      time.Duration
}

// PhaseOptions is a declarative form of Options for the common case where
// readiness is determined by a single status-phase string. The pending and
// failed phase sets are typically sourced from the resource's OpenAPI status
// enum (see cli-terraform-generator AsyncConfig).
type PhaseOptions[T any] struct {
	Fetch         func() (T, error)
	Phase         func(T) string
	PendingPhases []string
	FailedPhases  []string
	Initial       T
	Delay         time.Duration
	PollInterval  time.Duration
	Timeout       time.Duration
}

var (
	ErrTimeout             = errors.New("wait timed out")
	ErrFailed              = errors.New("resource entered failed state")
	ErrInvalidPollInterval = errors.New("poll interval must be greater than zero")
)

// PhaseSet returns a predicate reporting whether its argument is in phases.
// Used to build IsTerminal / IsFailed checks from OpenAPI status enums.
func PhaseSet(phases ...string) func(string) bool {
	set := make(map[string]struct{}, len(phases))
	for _, p := range phases {
		set[p] = struct{}{}
	}
	return func(s string) bool {
		_, ok := set[s]
		return ok
	}
}

// PollPhases polls opts.Fetch until opts.Phase returns a value not in
// opts.PendingPhases. If the resulting phase is in opts.FailedPhases the
// return error is ErrFailed; otherwise it is treated as a successful
// terminal state. Timeout / context-cancellation / fetch-errors behave as in
// Poll.
func PollPhases[T any](ctx context.Context, opts PhaseOptions[T]) (T, error) {
	pending := PhaseSet(opts.PendingPhases...)
	failed := PhaseSet(opts.FailedPhases...)
	return Poll(ctx, Options[T]{
		Fetch:        opts.Fetch,
		IsTerminal:   func(v T) bool { return !pending(opts.Phase(v)) },
		IsFailed:     func(v T) bool { return failed(opts.Phase(v)) },
		Initial:      opts.Initial,
		Delay:        opts.Delay,
		PollInterval: opts.PollInterval,
		Timeout:      opts.Timeout,
	})
}

// Poll sleeps opts.Delay (if non-zero), then calls opts.Fetch, then every
// opts.PollInterval until IsFailed returns true (ErrFailed), IsTerminal
// returns true (success), opts.Timeout elapses, or ctx is cancelled.
//
// Fetch errors are treated as transient and do not abort polling: the loop
// continues, preserving the most recent successfully-fetched value as `last`.
// This matches the historical retry.Retry-based behavior callers relied on,
// where 429/5xx/network blips during polling were retried until timeout. If
// the timeout elapses while the most recent Fetch errored, that error is
// returned in place of ErrTimeout so the user sees the underlying cause.
//
// Every Fetch runs under the same deadline, through Call: a fetch that is still
// in flight when the timeout elapses is abandoned (its goroutine finishes on its
// own) and Poll returns ErrTimeout, so a slow GET can neither hold the caller
// past Timeout nor turn into a success after it.
func Poll[T any](ctx context.Context, opts Options[T]) (T, error) {
	last := opts.Initial
	var lastErr error

	if opts.PollInterval <= 0 {
		return last, ErrInvalidPollInterval
	}

	fetchCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	check := func() (bool, error) {
		v, ferr := Call(fetchCtx, opts.Fetch)
		if errors.Is(ferr, context.DeadlineExceeded) && fetchCtx.Err() != nil && ctx.Err() == nil {
			// The poll's own deadline cut the fetch off: report the timeout, keeping the
			// last-known value (and a prior fetch error, as the deadline branch below does).
			if lastErr != nil {
				return true, lastErr
			}
			return true, ErrTimeout
		}
		if ferr != nil {
			lastErr = ferr
			return false, nil
		}
		lastErr = nil
		last = v
		if opts.IsFailed != nil && opts.IsFailed(v) {
			return true, ErrFailed
		}
		if opts.IsTerminal(v) {
			return true, nil
		}
		return false, nil
	}

	deadline := time.After(opts.Timeout)

	if opts.Delay > 0 {
		select {
		case <-time.After(opts.Delay):
		case <-deadline:
			return last, ErrTimeout
		case <-ctx.Done():
			return last, ctx.Err()
		}
	}

	if done, err := check(); done {
		return last, err
	}

	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if done, err := check(); done {
				return last, err
			}
		case <-deadline:
			if lastErr != nil {
				return last, lastErr
			}
			return last, ErrTimeout
		case <-ctx.Done():
			return last, ctx.Err()
		}
	}
}
