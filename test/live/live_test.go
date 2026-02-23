//go:build live_test

package live

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var liveBin = "test/live/bin/confluent"

// CLILiveTest represents a single step in a live integration test.
type CLILiveTest struct {
	// Name is the test step name shown in output.
	Name string
	// Args is the CLI arguments string (e.g., "environment create my-env -o json").
	// Supports {{.key}} template syntax when UseStateVars is true.
	Args string
	// ExitCode is the expected process exit code (default 0).
	ExitCode int
	// Input provides stdin content to the command.
	Input string
	// Contains lists strings that must appear in the command output.
	Contains []string
	// NotContains lists strings that must NOT appear in the command output.
	NotContains []string
	// Regex lists regex patterns that the output must match.
	Regex []string
	// JSONFields maps JSON field paths to expected values.
	// An empty string value means "any value" (just check field exists with a non-empty value).
	JSONFields map[string]string
	// JSONFieldsExist lists JSON fields that must exist in the output (value can be anything including empty).
	JSONFieldsExist []string
	// WantFunc is an optional custom assertion function.
	WantFunc func(t *testing.T, output string, state *LiveTestState)
	// CaptureID stores the extracted "id" field from JSON output into state under this key.
	CaptureID string
	// UseStateVars enables {{.key}} template substitution in Args from state.
	UseStateVars bool
	// Retries is the number of times to retry on failure (0 means no retry).
	// Useful for steps that may hit eventual consistency delays.
	Retries int
	// RetryInterval is the time to wait between retries. Defaults to 5s if Retries > 0.
	RetryInterval time.Duration
}

// LiveTestState is a thread-safe map for passing dynamic values between test steps.
// Each test gets its own LiveTestState with an isolated homeDir for CLI config.
type LiveTestState struct {
	mu      sync.Mutex
	values  map[string]string
	homeDir string
}

func NewLiveTestState() *LiveTestState {
	return &LiveTestState{values: make(map[string]string)}
}

func (s *LiveTestState) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

func (s *LiveTestState) Get(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.values[key]
}

func (s *LiveTestState) All() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make(map[string]string, len(s.values))
	for k, v := range s.values {
		cp[k] = v
	}
	return cp
}

// CLILiveTestSuite is the live integration test suite.
type CLILiveTestSuite struct {
	suite.Suite
	binPath string
}

func TestLive(t *testing.T) {
	if os.Getenv("CLI_LIVE_TEST") != "1" {
		t.Skip("Skipping live test. Set CLI_LIVE_TEST=1 to run.")
	}
	suite.Run(t, new(CLILiveTestSuite))
}

func (s *CLILiveTestSuite) SetupSuite() {
	req := require.New(s.T())

	// Navigate to repo root (test/live -> repo root is ../..)
	err := os.Chdir("../..")
	req.NoError(err)

	dir, err := os.Getwd()
	req.NoError(err)

	bin := liveBin
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	s.binPath = filepath.Join(dir, bin)
	_, err = os.Stat(s.binPath)
	req.NoError(err, "CLI binary not found at %s — run 'make build-for-live-test' first", s.binPath)
}

// setupTestContext creates an isolated CLI config directory and logs in.
// Each test gets its own HOME so tests can run concurrently without clobbering shared state.
func (s *CLILiveTestSuite) setupTestContext(t *testing.T) *LiveTestState {
	t.Helper()

	homeDir, err := os.MkdirTemp("", "cli-live-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(homeDir) })

	state := NewLiveTestState()
	state.homeDir = homeDir

	email := requiredEnv(t, "CONFLUENT_CLOUD_EMAIL")
	password := requiredEnv(t, "CONFLUENT_CLOUD_PASSWORD")

	output := s.runRawCommand(t, "login", []string{
		homeEnvVar(homeDir),
		fmt.Sprintf("CONFLUENT_CLOUD_EMAIL=%s", email),
		fmt.Sprintf("CONFLUENT_CLOUD_PASSWORD=%s", password),
	}, "", 0)
	require.Contains(t, output, "Logged in", "login failed: %s", output)

	return state
}

// homeEnvVar returns the HOME (or USERPROFILE on Windows) env var assignment for the given dir.
func homeEnvVar(dir string) string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE=" + dir
	}
	return "HOME=" + dir
}

// buildCommandEnv creates a command environment from os.Environ() with extraEnv overrides.
// Duplicate keys in os.Environ() are replaced by values from extraEnv.
func buildCommandEnv(extraEnv []string) []string {
	overrideKeys := make(map[string]bool)
	for _, e := range extraEnv {
		if idx := strings.IndexByte(e, '='); idx >= 0 {
			overrideKeys[e[:idx]] = true
		}
	}

	var result []string
	for _, e := range os.Environ() {
		if idx := strings.IndexByte(e, '='); idx >= 0 {
			if overrideKeys[e[:idx]] {
				continue
			}
		}
		result = append(result, e)
	}
	return append(result, extraEnv...)
}

// runLiveCommand executes a CLILiveTest step, performs template substitution, runs the command,
// validates output, and captures IDs. If the step has Retries > 0, the command is retried
// on failure with the specified interval between attempts.
func (s *CLILiveTestSuite) runLiveCommand(t *testing.T, step CLILiveTest, state *LiveTestState) string {
	t.Helper()

	args := step.Args
	if step.UseStateVars {
		args = substituteStateVars(t, args, state)
	}

	env := []string{homeEnvVar(state.homeDir)}

	var output string
	if step.Retries > 0 {
		interval := step.RetryInterval
		if interval == 0 {
			interval = 5 * time.Second
		}
		var lastErr error
		for attempt := 0; attempt <= step.Retries; attempt++ {
			if attempt > 0 {
				t.Logf("Retry %d/%d after %s", attempt, step.Retries, interval)
				time.Sleep(interval)
			}
			output, lastErr = s.tryRunRawCommand(args, env, step.Input, step.ExitCode)
			if lastErr == nil {
				break
			}
		}
		require.NoError(t, lastErr, "command 'confluent %s' failed after %d retries:\n%s", args, step.Retries, output)
	} else {
		output = s.runRawCommand(t, args, env, step.Input, step.ExitCode)
	}

	if step.CaptureID != "" {
		id := extractID(t, output)
		require.NotEmpty(t, id, "failed to extract ID from output for key %q:\n%s", step.CaptureID, output)
		state.Set(step.CaptureID, id)
		t.Logf("Captured %s = %s", step.CaptureID, id)
	}

	s.validateLiveOutput(t, step, output, state)

	return output
}

// runRawCommand executes the CLI binary with the given args string and returns combined output.
func (s *CLILiveTestSuite) runRawCommand(t *testing.T, argString string, env []string, input string, exitCode int) string {
	t.Helper()

	args := shellSplit(argString)
	cmd := exec.Command(s.binPath, args...)
	cmd.Env = buildCommandEnv(env)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	out, err := cmd.CombinedOutput()
	if exitCode == 0 {
		require.NoError(t, err, "command 'confluent %s' failed:\n%s", argString, string(out))
	}
	require.Equal(t, exitCode, cmd.ProcessState.ExitCode(),
		"unexpected exit code for 'confluent %s':\n%s", argString, string(out))

	return string(out)
}

// tryRunRawCommand is like runRawCommand but returns an error instead of failing the test.
// Used by runLiveCommand to support retries on transient failures.
func (s *CLILiveTestSuite) tryRunRawCommand(argString string, env []string, input string, expectedExitCode int) (string, error) {
	args := shellSplit(argString)
	cmd := exec.Command(s.binPath, args...)
	cmd.Env = buildCommandEnv(env)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	out, err := cmd.CombinedOutput()
	output := string(out)

	if expectedExitCode == 0 && err != nil {
		return output, fmt.Errorf("command 'confluent %s' failed:\n%s", argString, output)
	}
	if cmd.ProcessState.ExitCode() != expectedExitCode {
		return output, fmt.Errorf("unexpected exit code %d (expected %d) for 'confluent %s':\n%s",
			cmd.ProcessState.ExitCode(), expectedExitCode, argString, output)
	}

	return output, nil
}

// validateLiveOutput runs all assertion strategies on the command output.
func (s *CLILiveTestSuite) validateLiveOutput(t *testing.T, step CLILiveTest, output string, state *LiveTestState) {
	t.Helper()

	for _, c := range step.Contains {
		require.Contains(t, output, c, "output of '%s' missing expected string %q", step.Name, c)
	}

	for _, nc := range step.NotContains {
		require.NotContains(t, output, nc, "output of '%s' contains unexpected string %q", step.Name, nc)
	}

	for _, pattern := range step.Regex {
		re, err := regexp.Compile(pattern)
		require.NoError(t, err, "invalid regex pattern: %s", pattern)
		require.True(t, re.MatchString(output), "output of '%s' does not match regex %q:\n%s", step.Name, pattern, output)
	}

	if len(step.JSONFields) > 0 || len(step.JSONFieldsExist) > 0 {
		trimmed := strings.TrimSpace(output)
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
			// Try as JSON array and use the first element
			var arr []map[string]interface{}
			err2 := json.Unmarshal([]byte(trimmed), &arr)
			require.NoError(t, err2, "output of '%s' is not valid JSON:\n%s", step.Name, output)
			require.NotEmpty(t, arr, "JSON array output of '%s' is empty", step.Name)
			parsed = arr[0]
		}

		for field, expected := range step.JSONFields {
			val, ok := parsed[field]
			require.True(t, ok, "JSON output of '%s' missing field %q", step.Name, field)
			if expected != "" {
				require.Equal(t, expected, fmt.Sprintf("%v", val),
					"JSON field %q of '%s' has unexpected value", field, step.Name)
			} else {
				require.NotEmpty(t, fmt.Sprintf("%v", val),
					"JSON field %q of '%s' is empty", field, step.Name)
			}
		}

		for _, field := range step.JSONFieldsExist {
			_, ok := parsed[field]
			require.True(t, ok, "JSON output of '%s' missing field %q", step.Name, field)
		}
	}

	if step.WantFunc != nil {
		step.WantFunc(t, output, state)
	}
}

// registerCleanup schedules a CLI delete command to run during test cleanup (even on failure).
func (s *CLILiveTestSuite) registerCleanup(t *testing.T, deleteArgString string, state *LiveTestState) {
	t.Cleanup(func() {
		args := substituteStateVars(t, deleteArgString, state)
		t.Logf("Cleanup: confluent %s", args)
		// Best-effort cleanup: log but don't fail
		cmd := exec.Command(s.binPath, shellSplit(args)...)
		cmd.Env = buildCommandEnv([]string{homeEnvVar(state.homeDir)})
		if err := cmd.Run(); err != nil {
			t.Logf("Cleanup: resource already deleted, skipping")
		}
	})
}

// substituteStateVars replaces {{.key}} placeholders in a string with values from state.
func substituteStateVars(t *testing.T, s string, state *LiveTestState) string {
	t.Helper()

	if !strings.Contains(s, "{{") {
		return s
	}

	tmpl, err := template.New("args").Option("missingkey=error").Parse(s)
	require.NoError(t, err, "failed to parse template: %s", s)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, state.All())
	require.NoError(t, err, "failed to execute template: %s", s)

	return buf.String()
}

// extractID extracts the "id" field from JSON output.
func extractID(t *testing.T, output string) string {
	t.Helper()

	trimmed := strings.TrimSpace(output)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		// Not JSON — try to extract from table-style output (first column of second line)
		lines := strings.Split(trimmed, "\n")
		for i, line := range lines {
			if i == 0 {
				continue // skip header
			}
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0]
			}
		}
		return ""
	}

	if id, ok := parsed["id"]; ok {
		return fmt.Sprintf("%v", id)
	}
	return ""
}
