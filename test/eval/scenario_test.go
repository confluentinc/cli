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
	fakeRun := func(bin string, env []string, args string) error {
		mu.Lock()
		order = append(order, args)
		mu.Unlock()
		return nil
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
