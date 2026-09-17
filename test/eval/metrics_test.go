//go:build eval

package eval

import (
	"math"
	"testing"
)

func writeSessionConfig(t *testing.T, home, activeEnv string) {
	t.Helper()
	body := `{"current_context": "ctx-1", "contexts": {"ctx-1": {"current_environment": "` + activeEnv + `"}}}`
	writeConfig(t, home, body)
}

func TestGradeSessionOKWhenObservedMatchesIntended(t *testing.T) {
	home := t.TempDir()
	writeSessionConfig(t, home, "env-596")
	r := SessionResult{
		Session:     0,
		HomeDir:     home,
		IntendedEnv: "env-596",
		Invocations: []Invocation{{Command: "login --url http://mock", ExitCode: 0}},
	}

	got := GradeSession(r)

	if got.Verdict != VerdictOK {
		t.Fatalf("verdict = %q, want ok", got.Verdict)
	}
	if got.ObservedEnv != "env-596" {
		t.Fatalf("observed env = %q, want env-596", got.ObservedEnv)
	}
	if got.Detail != "" {
		t.Fatalf("detail = %q, want empty for ok verdict", got.Detail)
	}
}

func TestGradeSessionCollisionWhenObservedDiffersFromIntended(t *testing.T) {
	home := t.TempDir()
	writeSessionConfig(t, home, "env-595") // clobbered by a peer session

	got := GradeSession(SessionResult{Session: 1, HomeDir: home, IntendedEnv: "env-596"})

	if got.Verdict != VerdictCollision {
		t.Fatalf("verdict = %q, want collision", got.Verdict)
	}
	if got.ObservedEnv != "env-595" {
		t.Fatalf("observed env = %q, want env-595", got.ObservedEnv)
	}
	if got.Detail == "" {
		t.Fatalf("expected a non-empty detail explaining the collision")
	}
}

func TestGradeSessionCorruptionWhenConfigUnparseable(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context": "ctx-1", "contexts": {`) // torn write

	got := GradeSession(SessionResult{Session: 0, HomeDir: home, IntendedEnv: "env-596"})

	if got.Verdict != VerdictCorruption {
		t.Fatalf("verdict = %q, want corruption", got.Verdict)
	}
	if got.ObservedEnv != "" {
		t.Fatalf("observed env = %q, want empty on corruption", got.ObservedEnv)
	}
}

func TestGradeSessionErrorWhenInvocationFailedDoesNotCountAsCollision(t *testing.T) {
	home := t.TempDir()
	// config genuinely disagrees with intent - without the error short-circuit this would grade as
	// a real collision, so this proves error wins priority rather than merely absence of "ok".
	writeSessionConfig(t, home, "env-595")

	got := GradeSession(SessionResult{
		Session:     0,
		HomeDir:     home,
		IntendedEnv: "env-596",
		Invocations: []Invocation{{Command: "login --url http://mock", ExitCode: 1, Stderr: "connection refused"}},
	})

	if got.Verdict != VerdictError {
		t.Fatalf("verdict = %q, want error", got.Verdict)
	}
	if got.ObservedEnv != "" {
		t.Fatalf("observed env = %q, want empty on error (never got to grade the config)", got.ObservedEnv)
	}
	if got.Detail == "" {
		t.Fatalf("expected a non-empty detail explaining the failing invocation")
	}
}

func TestGradeTrialAllPassedOnlyWhenEverySessionOK(t *testing.T) {
	home1, home2 := t.TempDir(), t.TempDir()
	writeSessionConfig(t, home1, "env-596")
	writeSessionConfig(t, home2, "env-595") // collides with its own intent below

	trial := GradeTrial(0, []SessionResult{
		{Session: 0, HomeDir: home1, IntendedEnv: "env-596"},
		{Session: 1, HomeDir: home2, IntendedEnv: "env-596"},
	})

	if trial.AllPassed {
		t.Fatalf("expected AllPassed = false when one session collided")
	}
	if trial.Sessions[0].Verdict != VerdictOK || trial.Sessions[1].Verdict != VerdictCollision {
		t.Fatalf("unexpected verdicts: %+v", trial.Sessions)
	}
}

func TestAggregateComputesRatesIncludingErrorRate(t *testing.T) {
	trials := []TrialResult{
		{
			Trial: 0,
			Sessions: []SessionOutcome{
				{Verdict: VerdictOK},
				{Verdict: VerdictCollision},
			},
			AllPassed: false,
		},
		{
			Trial: 1,
			Sessions: []SessionOutcome{
				{Verdict: VerdictOK},
				{Verdict: VerdictError},
			},
			AllPassed: false,
		},
	}

	m := Aggregate(trials)

	if m.Trials != 2 {
		t.Fatalf("trials = %d, want 2", m.Trials)
	}
	if want := 1.0 / 4.0; math.Abs(m.CollisionRate-want) > 1e-9 {
		t.Fatalf("collision rate = %v, want %v", m.CollisionRate, want)
	}
	if want := 1.0 / 4.0; math.Abs(m.ErrorRate-want) > 1e-9 {
		t.Fatalf("error rate = %v, want %v", m.ErrorRate, want)
	}
	if m.CorruptionRate != 0 {
		t.Fatalf("corruption rate = %v, want 0", m.CorruptionRate)
	}
	if m.PassCaretK != 0 {
		t.Fatalf("pass^k = %v, want 0 (no trial had every session ok)", m.PassCaretK)
	}
}

func TestAggregateAllOKTrialsPassCaretKIsOne(t *testing.T) {
	trials := []TrialResult{
		{Sessions: []SessionOutcome{{Verdict: VerdictOK}, {Verdict: VerdictOK}}, AllPassed: true},
		{Sessions: []SessionOutcome{{Verdict: VerdictOK}, {Verdict: VerdictOK}}, AllPassed: true},
	}

	m := Aggregate(trials)

	if m.PassCaretK != 1.0 {
		t.Fatalf("pass^k = %v, want 1.0", m.PassCaretK)
	}
	if m.CollisionRate != 0 || m.CorruptionRate != 0 || m.ErrorRate != 0 {
		t.Fatalf("expected all rates 0, got %+v", m)
	}
}
