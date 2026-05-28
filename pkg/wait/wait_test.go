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

// TestPoll_PersistentFetchErrorReturnsLastErrAtTimeout: fetch errors do not
// abort polling (preserves the historical retry.Retry behavior); when the
// timeout fires while the most recent fetch errored, the underlying error is
// surfaced in place of ErrTimeout so users see the real cause.
func TestPoll_PersistentFetchErrorReturnsLastErrAtTimeout(t *testing.T) {
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
		Tick:       time.Millisecond,
		Timeout:    20 * time.Millisecond,
	})
	require.ErrorIs(t, err, fetchErr)
	require.NotErrorIs(t, err, ErrTimeout)
	require.Equal(t, "PENDING", v.phase) // last good value preserved
	require.GreaterOrEqual(t, calls, 2)
}

func TestPoll_FetchErrorOnlyOnFirstCallReturnsAtTimeout(t *testing.T) {
	fetchErr := fmt.Errorf("initial fetch failure")
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			return fakeResource{}, fetchErr
		},
		IsTerminal: func(r fakeResource) bool { return true },
		Tick:       time.Millisecond,
		Timeout:    20 * time.Millisecond,
	})
	require.ErrorIs(t, err, fetchErr)
	require.Equal(t, fakeResource{}, v)
}

// TestPoll_FetchErrorThenSuccess: a transient fetch error mid-polling should
// not abort. Once a subsequent fetch returns successfully and reaches a
// terminal state, the poll returns success.
func TestPoll_FetchErrorThenSuccess(t *testing.T) {
	calls := 0
	fetchErr := fmt.Errorf("transient 502")
	v, err := Poll(context.Background(), Options[fakeResource]{
		Fetch: func() (fakeResource, error) {
			calls++
			switch calls {
			case 1:
				return fakeResource{phase: "PENDING"}, nil
			case 2, 3:
				return fakeResource{}, fetchErr
			default:
				return fakeResource{phase: "READY"}, nil
			}
		},
		IsTerminal: func(r fakeResource) bool { return r.phase == "READY" },
		Tick:       time.Millisecond,
		Timeout:    time.Second,
	})
	require.NoError(t, err)
	require.Equal(t, "READY", v.phase)
	require.GreaterOrEqual(t, calls, 4)
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

func TestPhaseSet(t *testing.T) {
	pending := PhaseSet("PENDING", "FAILING", "STOPPING", "DELETING")
	require.True(t, pending("PENDING"))
	require.True(t, pending("FAILING"))
	require.True(t, pending("STOPPING"))
	require.True(t, pending("DELETING"))
	require.False(t, pending("RUNNING"))
	require.False(t, pending("FAILED"))
	require.False(t, pending(""))
	require.False(t, pending("pending")) // case-sensitive

	// Empty set never matches.
	none := PhaseSet()
	require.False(t, none(""))
	require.False(t, none("ANYTHING"))
}

func TestPollPhases_TerminalSuccess(t *testing.T) {
	calls := 0
	v, err := PollPhases(context.Background(), PhaseOptions[fakeResource]{
		Fetch: func() (fakeResource, error) {
			calls++
			if calls < 3 {
				return fakeResource{phase: "PENDING"}, nil
			}
			return fakeResource{phase: "RUNNING"}, nil
		},
		Phase:         func(r fakeResource) string { return r.phase },
		PendingPhases: []string{"PENDING", "FAILING", "STOPPING", "DELETING"},
		FailedPhases:  []string{"FAILED"},
		Tick:          time.Nanosecond,
		Timeout:       time.Second,
	})
	require.NoError(t, err)
	require.Equal(t, "RUNNING", v.phase)
	require.Equal(t, 3, calls)
}

func TestPollPhases_FailedPhase(t *testing.T) {
	v, err := PollPhases(context.Background(), PhaseOptions[fakeResource]{
		Fetch:         func() (fakeResource, error) { return fakeResource{phase: "FAILED"}, nil },
		Phase:         func(r fakeResource) string { return r.phase },
		PendingPhases: []string{"PENDING"},
		FailedPhases:  []string{"FAILED"},
		Tick:          time.Nanosecond,
		Timeout:       time.Second,
	})
	require.ErrorIs(t, err, ErrFailed)
	require.Equal(t, "FAILED", v.phase)
}

// Verifies the bug Channing flagged: PENDING is not the only intermediate
// state — FAILING, STOPPING, DELETING are also transitioning. They must keep
// polling, not terminate.
func TestPollPhases_AllPendingPhasesContinuePolling(t *testing.T) {
	cases := []struct {
		name      string
		sequence  []string
		wantFinal string
	}{
		{name: "failing_then_failed", sequence: []string{"PENDING", "FAILING", "FAILED"}, wantFinal: "FAILED"},
		{name: "stopping_then_stopped", sequence: []string{"PENDING", "STOPPING", "STOPPED"}, wantFinal: "STOPPED"},
		{name: "deleting_then_running", sequence: []string{"PENDING", "DELETING", "RUNNING"}, wantFinal: "RUNNING"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			v, err := PollPhases(context.Background(), PhaseOptions[fakeResource]{
				Fetch: func() (fakeResource, error) {
					p := tc.sequence[calls]
					calls++
					return fakeResource{phase: p}, nil
				},
				Phase:         func(r fakeResource) string { return r.phase },
				PendingPhases: []string{"PENDING", "FAILING", "STOPPING", "DELETING"},
				FailedPhases:  []string{"FAILED"},
				Tick:          time.Nanosecond,
				Timeout:       time.Second,
			})
			require.Equal(t, tc.wantFinal, v.phase)
			if tc.wantFinal == "FAILED" {
				require.ErrorIs(t, err, ErrFailed)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, len(tc.sequence), calls)
		})
	}
}
