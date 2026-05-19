package wait

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeResource struct {
	phase string
}

func TestPoll_ImmediateReady(t *testing.T) {
	calls := 0
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			calls++
			return fakeResource{phase: "READY"}, nil
		},
		IsTerminal: func(r fakeResource) bool { return r.phase == "READY" },
		Tick:       time.Millisecond,
		Timeout:    time.Second,
	})
	require.NoError(t, err)
	require.Equal(t, "READY", v.phase)
	require.Equal(t, 1, calls)
}

func TestPoll_EventuallyReady(t *testing.T) {
	calls := 0
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			calls++
			if calls < 3 {
				return fakeResource{phase: "PENDING"}, nil
			}
			return fakeResource{phase: "READY"}, nil
		},
		IsTerminal: func(r fakeResource) bool { return r.phase != "PENDING" },
		Tick:       time.Nanosecond,
		Timeout:    time.Second,
	})
	require.NoError(t, err)
	require.Equal(t, "READY", v.phase)
	require.Equal(t, 3, calls)
}

func TestPoll_Failed(t *testing.T) {
	calls := 0
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			calls++
			if calls == 1 {
				return fakeResource{phase: "PENDING"}, nil
			}
			return fakeResource{phase: "FAILED"}, nil
		},
		IsTerminal: func(r fakeResource) bool { return r.phase == "READY" || r.phase == "FAILED" },
		IsFailed:   func(r fakeResource) bool { return r.phase == "FAILED" },
		Tick:       time.Nanosecond,
		Timeout:    time.Second,
	})
	require.ErrorIs(t, err, ErrFailed)
	require.Equal(t, "FAILED", v.phase)
}

func TestPoll_Timeout(t *testing.T) {
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			return fakeResource{phase: "PENDING"}, nil
		},
		IsTerminal: func(r fakeResource) bool { return r.phase != "PENDING" },
		Tick:       time.Millisecond,
		Timeout:    5 * time.Millisecond,
	})
	require.ErrorIs(t, err, ErrTimeout)
	require.Equal(t, "PENDING", v.phase)
}

func TestPoll_FetchError(t *testing.T) {
	calls := 0
	fetchErr := fmt.Errorf("transient fetch failure")
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			calls++
			if calls == 1 {
				return fakeResource{phase: "PENDING"}, nil
			}
			return fakeResource{}, fetchErr
		},
		IsTerminal: func(r fakeResource) bool { return r.phase != "PENDING" },
		Tick:       time.Nanosecond,
		Timeout:    time.Second,
	})
	require.ErrorIs(t, err, fetchErr)
	require.Equal(t, "PENDING", v.phase)
}

func TestPoll_FetchErrorOnFirstCall(t *testing.T) {
	fetchErr := fmt.Errorf("initial fetch failure")
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			return fakeResource{}, fetchErr
		},
		IsTerminal: func(r fakeResource) bool { return true },
		Tick:       time.Nanosecond,
		Timeout:    time.Second,
	})
	require.ErrorIs(t, err, fetchErr)
	require.Equal(t, fakeResource{}, v)
}

func TestPoll_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	_, err := Poll(ctx, Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			return fakeResource{phase: "PENDING"}, nil
		},
		IsTerminal: func(r fakeResource) bool { return r.phase != "PENDING" },
		Tick:       time.Millisecond,
		Timeout:    time.Second,
	})
	require.True(t, errors.Is(err, context.Canceled))
}
