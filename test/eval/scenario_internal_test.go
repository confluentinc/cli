//go:build eval

package eval

import (
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

func TestLoginStepCarriesURL(t *testing.T) {
	if got := loginStep("http://127.0.0.1:9000"); got != "login --url http://127.0.0.1:9000" {
		t.Errorf("loginStep built %q", got)
	}
}
