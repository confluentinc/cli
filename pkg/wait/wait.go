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

var (
	ErrTimeout = errors.New("wait timed out")
	ErrFailed  = errors.New("resource entered failed state")
)

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
