//go:build eval

package eval

import (
	"math"
	"testing"
)

func TestGradeSessionOKWhenObservedMatchesIntended(t *testing.T) {
	home := t.TempDir()
	writeEnvConfig(t, home, "ctx-1", "env-596")
	r := SessionResult{
		Session:     0,
		HomeDir:     home,
		Invocations: []Invocation{{Command: "login --url http://mock", ExitCode: 0}},
	}

	got := GradeSession(r, "env-596", observeEnv)

	if got.Verdict != VerdictOK {
		t.Fatalf("verdict = %q, want ok", got.Verdict)
	}
	if got.Observed != "env-596" {
		t.Fatalf("observed env = %q, want env-596", got.Observed)
	}
	if got.Detail != "" {
		t.Fatalf("detail = %q, want empty for ok verdict", got.Detail)
	}
}

func TestGradeSessionCollisionWhenObservedDiffersFromIntended(t *testing.T) {
	home := t.TempDir()
	writeEnvConfig(t, home, "ctx-1", "env-595") // clobbered by a peer session

	got := GradeSession(SessionResult{Session: 1, HomeDir: home}, "env-596", observeEnv)

	if got.Verdict != VerdictCollision {
		t.Fatalf("verdict = %q, want collision", got.Verdict)
	}
	if got.Observed != "env-595" {
		t.Fatalf("observed env = %q, want env-595", got.Observed)
	}
	if got.Detail == "" {
		t.Fatalf("expected a non-empty detail explaining the collision")
	}
}

func TestGradeSessionCorruptionWhenConfigUnparseable(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context": "ctx-1", "contexts": {`) // torn write

	got := GradeSession(SessionResult{Session: 0, HomeDir: home}, "env-596", observeEnv)

	if got.Verdict != VerdictCorruption {
		t.Fatalf("verdict = %q, want corruption", got.Verdict)
	}
	if got.Observed != "" {
		t.Fatalf("observed env = %q, want empty on corruption", got.Observed)
	}
}

func TestGradeSessionErrorWhenInvocationFailedDoesNotCountAsCollision(t *testing.T) {
	home := t.TempDir()
	// config genuinely disagrees with intent - without the error short-circuit this would grade as
	// a real collision, so this proves error wins priority rather than merely absence of "ok".
	writeEnvConfig(t, home, "ctx-1", "env-595")

	got := GradeSession(SessionResult{
		Session:     0,
		HomeDir:     home,
		Invocations: []Invocation{{Command: "login --url http://mock", ExitCode: 1, Stderr: "connection refused"}},
	}, "env-596", observeEnv)

	if got.Verdict != VerdictError {
		t.Fatalf("verdict = %q, want error", got.Verdict)
	}
	if got.Observed != "" {
		t.Fatalf("observed env = %q, want empty on error (never got to grade the config)", got.Observed)
	}
	if got.Detail == "" {
		t.Fatalf("expected a non-empty detail explaining the failing invocation")
	}
}

func TestGradeTrialAllPassedOnlyWhenEverySessionOK(t *testing.T) {
	home1, home2 := t.TempDir(), t.TempDir()
	// The grade closure always grades against an empty intent, so an empty observed env is what
	// "passes" here; a non-empty one collides.
	writeEnvConfig(t, home1, "ctx-1", "")
	writeEnvConfig(t, home2, "ctx-1", "env-595")

	grade := func(results []SessionResult) []SessionOutcome {
		outs := make([]SessionOutcome, len(results))
		for i, r := range results {
			outs[i] = GradeSession(r, "", func(r SessionResult) (string, error) { return ReadActiveEnvironment(r.HomeDir) })
		}
		return outs
	}
	trial := GradeTrial(0, []SessionResult{
		{Session: 0, HomeDir: home1},
		{Session: 1, HomeDir: home2},
	}, grade)

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

func TestAssertRedSharedGreenIsolated(t *testing.T) {
	damaged := CellMetrics{CollisionRate: 0.5}
	clean := CellMetrics{PassCaretK: 1.0}
	tests := []struct {
		name             string
		shared, isolated CellMetrics
		wantViolations   int
	}{
		{"red shared, green isolated", damaged, clean, 0},
		{"corruption alone counts as shared damage", CellMetrics{CorruptionRate: 0.2}, clean, 0},
		{"clean shared is not a demonstration", CellMetrics{}, clean, 1},
		{"error-only shared is not a demonstration", CellMetrics{ErrorRate: 1.0}, clean, 1},
		{"isolated collision", damaged, CellMetrics{CollisionRate: 0.1, PassCaretK: 1.0}, 1},
		{"isolated corruption", damaged, CellMetrics{CorruptionRate: 0.1, PassCaretK: 1.0}, 1},
		{"isolated error", damaged, CellMetrics{ErrorRate: 0.1, PassCaretK: 1.0}, 1},
		{"isolated flaky trial", damaged, CellMetrics{PassCaretK: 0.9}, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := AssertRedSharedGreenIsolated(tc.shared, tc.isolated)

			if len(v) != tc.wantViolations {
				t.Errorf("got %d violation(s) %q, want %d", len(v), v, tc.wantViolations)
			}
		})
	}
}
