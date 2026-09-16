//go:build eval

package eval

import (
	"sync"
	"testing"
)

func TestRunScenarioHoldsBarrierUntilAllSessionsWrote(t *testing.T) {
	root := t.TempDir()
	p := NewIsolatedProvisioner(root)
	sessions := []Session{{IntendedEnv: "env-596"}, {IntendedEnv: "env-595"}}

	var mu sync.Mutex
	var order []string
	fakeRun := func(bin string, env []string, args string) Invocation {
		mu.Lock()
		order = append(order, args)
		mu.Unlock()
		return Invocation{Command: args, ExitCode: 0}
	}

	results := RunScenario("confluent", "http://mock", p, sessions, fakeRun)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	// Both sessions must have logged in and used an environment.
	var useCount, loginCount int
	mu.Lock()
	for _, a := range order {
		switch {
		case a == "login --url http://mock":
			loginCount++
		case a == "environment use env-596" || a == "environment use env-595":
			useCount++
		}
	}
	mu.Unlock()
	if loginCount != 2 || useCount != 2 {
		t.Fatalf("expected 2 logins and 2 env-use calls, got %d and %d", loginCount, useCount)
	}
	for _, r := range results {
		if r.HomeDir == "" {
			t.Fatalf("session %d has empty HomeDir", r.Session)
		}
	}
}

func TestRunScenarioCapturesInvocationsInOrder(t *testing.T) {
	root := t.TempDir()
	p := NewIsolatedProvisioner(root)
	sessions := []Session{{IntendedEnv: "env-596"}}

	fakeRun := func(bin string, env []string, args string) Invocation {
		return Invocation{Command: args, ExitCode: 0, DurationMs: 7}
	}

	results := RunScenario("confluent", "http://mock", p, sessions, fakeRun)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	invocations := results[0].Invocations
	if len(invocations) != 2 {
		t.Fatalf("got %d invocations, want 2 (login, environment use)", len(invocations))
	}
	if invocations[0].Command != "login --url http://mock" {
		t.Fatalf("invocation[0].Command = %q, want login first", invocations[0].Command)
	}
	if invocations[1].Command != "environment use env-596" {
		t.Fatalf("invocation[1].Command = %q, want environment use second", invocations[1].Command)
	}
	for _, inv := range invocations {
		if inv.ExitCode != 0 || inv.DurationMs != 7 {
			t.Fatalf("invocation not captured faithfully: %+v", inv)
		}
	}
}

func TestRunScenarioStopsSessionOnFailedInvocation(t *testing.T) {
	root := t.TempDir()
	p := NewIsolatedProvisioner(root)
	sessions := []Session{{IntendedEnv: "env-596"}}

	fakeRun := func(bin string, env []string, args string) Invocation {
		return Invocation{Command: args, ExitCode: 1, Stderr: "boom"}
	}

	results := RunScenario("confluent", "http://mock", p, sessions, fakeRun)

	if len(results[0].Invocations) != 1 {
		t.Fatalf("expected session to stop after the failing login, got %d invocations", len(results[0].Invocations))
	}
}
