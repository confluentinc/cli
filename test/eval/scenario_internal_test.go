//go:build eval

package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunScenarioBarrierReleasesContendAfterAllSetup(t *testing.T) {
	// Arrange: session 0 has a slow setup; without the barrier, session 1's contend
	// step would run before session 0 finished setup.
	root := t.TempDir()
	p := NewIsolatedProvisioner(root)
	scripts := []SessionScript{
		{Setup: []string{"setup slow"}, Contend: []string{"contend a"}},
		{Setup: []string{"setup fast"}, Contend: []string{"contend b"}},
	}
	var setupDone, violations atomic.Int32
	run := func(bin string, env []string, args string) Invocation {
		switch {
		case strings.HasPrefix(args, "setup"):
			if strings.Contains(args, "slow") {
				time.Sleep(50 * time.Millisecond)
			}
			setupDone.Add(1)
		default: // a contend step must not start until every setup step is done
			if setupDone.Load() != int32(len(scripts)) {
				violations.Add(1)
			}
		}
		return Invocation{Command: args, ExitCode: 0}
	}

	// Act
	results := RunScenario("bin", p, scripts, run)

	// Assert
	if violations.Load() != 0 {
		t.Errorf("a contend step ran before all setup completed (barrier not holding)")
	}
	for i, r := range results {
		if len(r.Invocations) != 2 {
			t.Errorf("session %d: expected 2 invocations, got %d", i, len(r.Invocations))
		}
	}
}

func TestRunScenarioHaltsSessionOnFailedSetupStep(t *testing.T) {
	// Arrange
	root := t.TempDir()
	p := NewIsolatedProvisioner(root)
	scripts := []SessionScript{
		{Setup: []string{"login --url x", "environment use env-a"}, Contend: []string{"kafka cluster use lkc-1"}},
	}
	run := func(bin string, env []string, args string) Invocation {
		if args == "login --url x" {
			return Invocation{Command: args, ExitCode: 1, Stderr: "boom"}
		}
		return Invocation{Command: args, ExitCode: 0}
	}

	// Act
	results := RunScenario("bin", p, scripts, run)

	// Assert: only the failed login ran; env-use and the contend step were skipped.
	if n := len(results[0].Invocations); n != 1 {
		t.Fatalf("expected the session to stop after the failed setup step, ran %d steps", n)
	}
}

// writeEnvConfig writes a config.json with the given context/environment. Named distinctly from
// grader_test.go's writeConfig(t, home, body), which takes a raw JSON body instead.
func writeEnvConfig(t *testing.T, home, ctx, env string) {
	t.Helper()
	dir := filepath.Join(home, ".confluent")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"current_context":%q,"contexts":{%q:{"current_environment":%q}}}`, ctx, ctx, env)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// observeEnv adapts the homeDir-based reader to the observe(SessionResult) shape.
func observeEnv(r SessionResult) (string, error) { return ReadActiveEnvironment(r.HomeDir) }

func TestGradeSessionLadder(t *testing.T) {
	home := t.TempDir()
	writeEnvConfig(t, home, "ctx", "env-a")

	ok := GradeSession(SessionResult{Session: 0, HomeDir: home}, "env-a", observeEnv)
	if ok.Verdict != VerdictOK || ok.Observed != "env-a" {
		t.Errorf("expected ok/env-a, got %s/%q", ok.Verdict, ok.Observed)
	}

	col := GradeSession(SessionResult{Session: 1, HomeDir: home}, "env-b", observeEnv)
	if col.Verdict != VerdictCollision {
		t.Errorf("expected collision, got %s", col.Verdict)
	}

	errRes := SessionResult{Session: 2, HomeDir: home, Invocations: []Invocation{{Command: "login", ExitCode: 1, Stderr: "no"}}}
	if e := GradeSession(errRes, "env-a", observeEnv); e.Verdict != VerdictError {
		t.Errorf("expected error to win the ladder, got %s", e.Verdict)
	}
}

func TestGradeSessionCorruptionOnBadConfig(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".confluent")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o644)

	if out := GradeSession(SessionResult{Session: 0, HomeDir: home}, "env-a", observeEnv); out.Verdict != VerdictCorruption {
		t.Errorf("expected corruption on unparseable config, got %s", out.Verdict)
	}
}

func TestGradeSessionCarriesLabelFromSessionResult(t *testing.T) {
	home := t.TempDir()
	writeEnvConfig(t, home, "ctx", "env-a")

	out := GradeSession(SessionResult{Session: 0, HomeDir: home, Label: "logs out"}, "env-a", observeEnv)

	if out.Label != "logs out" {
		t.Errorf("expected Label to carry through to the outcome, got %q", out.Label)
	}
}
