//go:build eval

package eval

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const (
	evalTrials = 5
	envA       = "env-596"
	envB       = "env-595"
)

// realRun executes one confluent invocation with the given per-session env and space-split args,
// bounded by a timeout so a hung subprocess can't hang the whole eval.
func realRun(bin string, env []string, args string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, splitArgs(args)...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &runError{args: args, out: string(out), err: err}
	}
	return nil
}

func TestEnvironmentCrosstalkEval(t *testing.T) {
	bin := buildCLI(t) // absolute path to test/bin/confluent
	backend := testserver.StartTestBackend(t, false)
	t.Cleanup(backend.Close)
	cloudURL := backend.GetCloudUrl()

	sessions := []Session{{IntendedEnv: envA}, {IntendedEnv: envB}}
	report := Report{Build: gitShortSHA(repoRootFromTest(t)), Cells: map[string]CellMetrics{}}

	for _, cell := range []struct {
		name string
		make func(root string) Provisioner
	}{
		{"shared", NewSharedProvisioner},
		{"isolated", NewIsolatedProvisioner},
	} {
		var outcomes []TrialOutcome
		for trial := 0; trial < evalTrials; trial++ {
			root := t.TempDir()
			p := cell.make(root)
			results := RunScenario(bin, cloudURL, p, sessions, realRun)
			for _, r := range results {
				if r.RunErr != nil {
					t.Logf("%s trial %d session %d run error: %v", cell.name, trial, r.Session, r.RunErr)
				}
			}
			outcomes = append(outcomes, GradeTrial(results))
		}
		report.Cells[cell.name] = Aggregate(outcomes)
	}

	t.Log("\n" + report.Summary())
	writeReport(t, report)

	shared := report.Cells["shared"]
	isolated := report.Cells["isolated"]

	// The headline claim: shared state collides or corrupts under concurrent writes, isolated
	// state does neither. A shared config.json can end up with the wrong active environment
	// (collision) or a torn/invalid write (corruption) - both are concurrent-state damage that
	// isolation removes.
	if shared.CollisionRate == 0 && shared.CorruptionRate == 0 {
		t.Errorf("expected collisions or corruptions under shared state, got both rates 0 (barrier or scenario broken?)")
	}
	if isolated.CollisionRate != 0 {
		t.Errorf("expected zero collisions under isolated state, got %.3f", isolated.CollisionRate)
	}
	if isolated.CorruptionRate != 0 {
		t.Errorf("expected zero corruptions under isolated state, got %.3f", isolated.CorruptionRate)
	}
	if isolated.PassCaretK != 1.0 {
		t.Errorf("expected isolated pass^k = 1.0, got %.3f", isolated.PassCaretK)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	// build-for-integration-test runs from the repo root and emits test/bin/confluent.
	repoRoot := repoRootFromTest(t)
	target := "build-for-integration-test"
	if runtime.GOOS == "windows" {
		target += "-windows"
	}
	cmd := exec.Command("make", target)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("make %s failed: %v\n%s", target, err, out)
	}
	bin := filepath.Join(repoRoot, "test", "bin", "confluent")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("built binary missing at %s: %v", bin, err)
	}
	return bin
}

type runError struct {
	args string
	out  string
	err  error
}

func (e *runError) Error() string {
	return e.args + ": " + e.err.Error() + "\n" + e.out
}

// splitArgs splits a command string on spaces. Phase-1 scenarios use no quoted/spaced arguments.
func splitArgs(s string) []string {
	return strings.Fields(s)
}

// gitShortSHA resolves the current checkout's short SHA so the report is self-labeling for
// cross-commit comparison. Falls back to "unknown" rather than failing the eval over a missing SHA.
func gitShortSHA(repoRoot string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func repoRootFromTest(t *testing.T) string {
	t.Helper()
	// tests run from test/eval/, so the repo root is two levels up.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func writeReport(t *testing.T, r Report) {
	t.Helper()
	dir := filepath.Join(repoRootFromTest(t), "test", "eval", "results")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := r.WriteJSON(filepath.Join(dir, "environment-crosstalk.json")); err != nil {
		t.Fatal(err)
	}
}
