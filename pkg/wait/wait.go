package wait

import (
	"context"
	"errors"
	"time"
)

// Options describes a single Poll invocation. T is the polled resource type.
type Options[T any] struct {
	Fetch      func() (T, error)
	IsTerminal func(T) bool
	IsFailed   func(T) bool
	Tick       time.Duration
	Timeout    time.Duration
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
	Tick          time.Duration
	Timeout       time.Duration
}

var (
	ErrTimeout = errors.New("wait timed out")
	ErrFailed  = errors.New("resource entered failed state")
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
		Fetch:      opts.Fetch,
		IsTerminal: func(v T) bool { return !pending(opts.Phase(v)) },
		IsFailed:   func(v T) bool { return failed(opts.Phase(v)) },
		Tick:       opts.Tick,
		Timeout:    opts.Timeout,
	})
}

// Poll calls opts.Fetch immediately, then every opts.Tick until IsFailed
// returns true (ErrFailed), IsTerminal returns true (success), opts.Timeout
// elapses (ErrTimeout), ctx is cancelled (ctx.Err()), or Fetch returns a
// non-nil error. Always returns the most recent successfully-fetched T.
func Poll[T any](ctx context.Context, opts Options[T]) (T, error) {
	var last T

	check := func() (T, bool, error) {
		v, err := opts.Fetch()
		if err != nil {
			return last, true, err
		}
		last = v
		if opts.IsFailed != nil && opts.IsFailed(v) {
			return last, true, ErrFailed
		}
		if opts.IsTerminal(v) {
			return last, true, nil
		}
		return last, false, nil
	}

	if v, done, err := check(); done {
		return v, err
	}

	ticker := time.NewTicker(opts.Tick)
	defer ticker.Stop()
	deadline := time.After(opts.Timeout)

	for {
		select {
		case <-ticker.C:
			if v, done, err := check(); done {
				return v, err
			}
		case <-deadline:
			return last, ErrTimeout
		case <-ctx.Done():
			return last, ctx.Err()
		}
	}
}
