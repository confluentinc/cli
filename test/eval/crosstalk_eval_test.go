//go:build eval

package eval

import (
	"bytes"
	"context"
	"errors"
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
// bounded by a timeout so a hung subprocess can't hang the whole eval, and captures the full
// transcript (exit code, stdout, stderr, duration) instead of collapsing it to a bare error.
func realRun(coverDir string) CommandFunc {
	return func(bin string, env []string, args string) Invocation {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, bin, splitArgs(args)...)
		// bin is built coverage-instrumented (build-for-integration-test), so it warns
		// "GOCOVERDIR not set" on stderr unless pointed somewhere; coverDir is a throwaway temp dir
		// that exists purely to silence that warning in captured transcripts. This eval's coverage is
		// intentionally NOT merged into `make coverage` - the eval package is excluded from the
		// normal suite by its `eval` build tag, so nothing else reads these coverage files.
		cmd.Env = append(append([]string{}, env...), "GOCOVERDIR="+coverDir)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		inv := Invocation{
			Command:    args,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
			DurationMs: time.Since(start).Milliseconds(),
		}
		if cmd.ProcessState != nil {
			inv.ExitCode = cmd.ProcessState.ExitCode()
		}
		switch {
		case ctx.Err() != nil:
			inv.Err = "timeout: " + ctx.Err().Error()
		case cmd.ProcessState == nil && err != nil:
			inv.Err = err.Error() // spawn failure - the process never ran
			inv.ExitCode = -1     // no real exit code exists; avoid rendering a misleading "exit 0"
		case err != nil && !errors.As(err, new(*exec.ExitError)):
			inv.Err = err.Error()
		}
		return inv
	}
}

func TestEnvironmentCrosstalkEval(t *testing.T) {
	bin := buildCLI(t) // absolute path to test/bin/confluent
	backend := testserver.StartTestBackend(t, false)
	t.Cleanup(backend.Close)
	cloudURL := backend.GetCloudUrl()

	sessions := []Session{{IntendedEnv: envA}, {IntendedEnv: envB}}
	run := realRun(t.TempDir()) // one shared GOCOVERDIR for every invocation in this test

	cells := []CellReport{}
	for _, cell := range []struct {
		name string
		make func(root string) Provisioner
	}{
		{"shared", NewSharedProvisioner},
		{"isolated", NewIsolatedProvisioner},
	} {
		var trials []TrialResult
		for trial := 0; trial < evalTrials; trial++ {
			root := t.TempDir()
			p := cell.make(root)
			results := RunScenario(bin, cloudURL, p, sessions, run)
			tr := GradeTrial(trial, results)
			for _, s := range tr.Sessions {
				if s.Verdict == VerdictError {
					t.Logf("%s trial %d session %d errored: %s", cell.name, trial, s.Session, s.Detail)
				}
			}
			trials = append(trials, tr)
		}
		cells = append(cells, CellReport{Name: cell.name, Metrics: Aggregate(trials), Trials: trials})
	}

	report := Report{
		Build:       gitShortSHA(repoRootFromTest(t)),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Scenarios: []ScenarioReport{
			{
				Name:        "environment-crosstalk",
				Description: "measures whether concurrent `confluent login` + `confluent environment use` sessions collide or corrupt state when sharing a config directory, versus each session getting its own.",
				Sessions:    len(sessions),
				Trials:      evalTrials,
				Cells:       cells,
			},
		},
	}

	t.Log("\n" + report.Summary())
	writeReport(t, report)

	shared := cellByName(cells, "shared").Metrics
	isolated := cellByName(cells, "isolated").Metrics

	// The headline claim: shared state collides or corrupts under concurrent writes; isolated state
	// does neither. A shared config.json can end up with the wrong active environment (collision) or
	// a torn/invalid write (corruption) - both are concurrent-state damage that isolation removes.
	// Invocation errors are graded separately (VerdictError) and do NOT prove crosstalk on their
	// own, so an error-only shared run must still fail this check rather than pass by coincidence.
	if shared.CollisionRate == 0 && shared.CorruptionRate == 0 {
		t.Errorf("expected state damage (collisions or corruptions) under shared state, got none - crosstalk not demonstrated (an error-only run does not prove the collision; barrier or scenario may be broken)")
	}
	if isolated.CollisionRate != 0 {
		t.Errorf("expected zero collisions under isolated state, got %.3f", isolated.CollisionRate)
	}
	if isolated.CorruptionRate != 0 {
		t.Errorf("expected zero corruptions under isolated state, got %.3f", isolated.CorruptionRate)
	}
	if isolated.ErrorRate != 0 {
		t.Errorf("expected zero errors under isolated state, got %.3f", isolated.ErrorRate)
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
	if err := r.WriteJSON(filepath.Join(dir, "report.json")); err != nil {
		t.Fatal(err)
	}
	if err := r.WriteHTMLReport(dir); err != nil {
		t.Fatal(err)
	}
}
