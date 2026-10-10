//go:build eval

package eval_test

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

	"github.com/confluentinc/cli/v4/test/eval"
	"github.com/confluentinc/cli/v4/test/eval/scenarios"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const evalTrials = 10

// realRun executes one confluent invocation with the given per-session env and space-split args,
// bounded by a timeout so a hung subprocess can't hang the whole eval, and captures the full
// transcript (exit code, stdout, stderr, duration) instead of collapsing it to a bare error.
func realRun(coverDir string) eval.CommandFunc {
	return func(bin string, env []string, args string) eval.Invocation {
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
		inv := eval.Invocation{
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

func TestEval(t *testing.T) {
	bin := buildCLI(t)
	backend := testserver.StartTestBackend(t, false)
	t.Cleanup(backend.Close)
	cloudURL := backend.GetCloudUrl()
	run := realRun(t.TempDir())

	var scenarioReports []eval.ScenarioReport
	for _, sc := range scenarios.All {
		scripts := sc.Sessions(cloudURL)
		var cells []eval.CellReport
		var sharedM, isolatedM eval.CellMetrics
		for _, cell := range []struct {
			name string
			make func(root string) eval.Provisioner
		}{{"shared", eval.NewSharedProvisioner}, {"isolated", eval.NewIsolatedProvisioner}} {
			var trials []eval.TrialResult
			for trial := 0; trial < evalTrials; trial++ {
				p := cell.make(t.TempDir())
				results := eval.RunScenario(bin, p, scripts, run)
				trials = append(trials, eval.GradeTrial(trial, results, sc.Grade))
			}
			m := eval.Aggregate(trials)
			cells = append(cells, eval.CellReport{Name: cell.name, Metrics: m, Trials: trials})
			if cell.name == "shared" {
				sharedM = m
			} else {
				isolatedM = m
			}
		}
		for _, v := range eval.AssertRedSharedGreenIsolated(sharedM, isolatedM) {
			t.Errorf("scenario %q: %s", sc.Name, v)
		}
		scenarioReports = append(scenarioReports, eval.ScenarioReport{
			Name: sc.Name, Description: sc.Description, Sessions: len(scripts), Trials: evalTrials, Cells: cells,
		})
	}

	report := eval.Report{Build: gitShortSHA(repoRootFromTest(t)), GeneratedAt: time.Now().UTC().Format(time.RFC3339), Scenarios: scenarioReports}
	t.Log("\n" + report.Summary())
	writeReport(t, report)
}

func buildCLI(t *testing.T) string {
	t.Helper()
	// Allow running against a binary built elsewhere (e.g. from another git ref) instead of the
	// current checkout - the seed of the v4-vs-v5 build dimension. Must be a coverage-instrumented
	// build-for-integration-test binary, or login --url won't take the cloud path.
	if bin := os.Getenv("EVAL_CLI_BIN"); bin != "" {
		if _, err := os.Stat(bin); err != nil {
			t.Fatalf("EVAL_CLI_BIN=%q not usable: %v", bin, err)
		}
		return bin
	}
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
	// Allow labeling the report with an explicit ref (e.g. the origin/main SHA when running against
	// an origin/main-built binary via EVAL_CLI_BIN) rather than the current checkout's HEAD.
	if label := os.Getenv("EVAL_BUILD_LABEL"); label != "" {
		return label
	}
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

func writeReport(t *testing.T, r eval.Report) {
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

func TestBuildCLIHonorsEvalCLIBinOverride(t *testing.T) {
	// Arrange: a fake prebuilt binary and the override pointing at it.
	dir := t.TempDir()
	fake := filepath.Join(dir, "confluent")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EVAL_CLI_BIN", fake)

	// Act
	got := buildCLI(t)

	// Assert: returned the override without building.
	if got != fake {
		t.Errorf("buildCLI returned %q, want the EVAL_CLI_BIN override %q", got, fake)
	}
}

func TestGitShortSHAHonorsBuildLabelOverride(t *testing.T) {
	t.Setenv("EVAL_BUILD_LABEL", "78f96cece")
	if got := gitShortSHA("/nonexistent-repo-root"); got != "78f96cece" {
		t.Errorf("gitShortSHA returned %q, want the EVAL_BUILD_LABEL override", got)
	}
}
