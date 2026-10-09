//go:build cuj_test

package cuj

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	claudeBin           = "test/cuj/agents/node_modules/.bin/claude"
	defaultAgentTimeout = 7 * time.Minute
)

// agentHiddenEnv are environment variables Claude must not see. The harness has
// already logged in with them, so the agent never needs the account's password.
var agentHiddenEnv = []string{"CONFLUENT_CLOUD_EMAIL", "CONFLUENT_CLOUD_PASSWORD"}

type agentResult struct {
	Transcript string
	DurationS  int
	ExitCode   int
	TimedOut   bool
}

// setupClaude checks that Claude Code was installed by `make install-cuj-agents`,
// so every run uses the version pinned in the Makefile.
func (s *CUJTestSuite) setupClaude() {
	t := s.T()
	path, err := filepath.Abs(claudeBin)
	require.NoError(t, err)
	_, err = os.Stat(path)
	require.NoError(t, err, "Claude Code not found at %s — run `make install-cuj-agents` first", path)

	out, err := exec.Command(path, "--version").Output()
	require.NoError(t, err, "`%s --version` failed", path)

	s.claudePath = path
	s.claudeVersion = strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	s.model = os.Getenv("CLI_CUJ_AGENT_MODEL")
}

// runAgent runs Claude Code headless (`claude -p`) with only the task. Claude
// runs in an empty working directory so it can't read this repo's docs, may use
// only the Bash tool, and finds the confluent wrapper (built binary + isolated
// HOME) first on PATH. HOME stays real so Claude's own login works. Claude runs
// in its own process group, which is killed when it finishes or times out, so
// commands it started (such as a consumer that never exits) don't outlive it.
func (s *CUJTestSuite) runAgent(t *testing.T, task, runDir string) agentResult {
	t.Helper()

	timeout := defaultAgentTimeout
	if v := os.Getenv("CLI_CUJ_AGENT_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		require.NoError(t, err, "invalid CLI_CUJ_AGENT_TIMEOUT %q (want a Go duration, e.g. 10m)", v)
		timeout = d
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{"-p", task, "--allowedTools", "Bash", "--output-format", "json"}
	if s.model != "" {
		args = append(args, "--model", s.model)
	}
	cmd := exec.CommandContext(ctx, s.claudePath, args...)
	cmd.Dir = t.TempDir()
	path := s.wrapperDir + string(os.PathListSeparator) + os.Getenv("PATH")
	cmd.Env = withEnv(withoutEnv(os.Environ(), agentHiddenEnv...), "PATH="+path, "DISABLE_AUTOUPDATER=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 10 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	t.Logf("running claude (timeout %s)...", timeout)
	start := time.Now()
	err := cmd.Run()
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	res := agentResult{
		Transcript: claudeTranscript(stdout.Bytes()),
		DurationS:  int(time.Since(start).Seconds()),
		TimedOut:   ctx.Err() == context.DeadlineExceeded,
	}
	require.NotNil(t, cmd.ProcessState, "failed to start %s: %v", s.claudePath, err)
	res.ExitCode = cmd.ProcessState.ExitCode()

	require.NoError(t, os.WriteFile(filepath.Join(runDir, "stdout.txt"), stdout.Bytes(), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "stderr.txt"), stderr.Bytes(), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "transcript.txt"), []byte(res.Transcript), 0o644))

	t.Logf("claude finished: exit=%d duration=%ds timed_out=%t", res.ExitCode, res.DurationS, res.TimedOut)
	if res.ExitCode != 0 && !res.TimedOut {
		t.Logf("claude stderr: %s", strings.TrimSpace(stderr.String()))
	}
	return res
}

// claudeTranscript extracts Claude's final answer from `--output-format json` output.
func claudeTranscript(stdout []byte) string {
	var envelope struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(stdout, &envelope); err == nil && envelope.Result != "" {
		return envelope.Result
	}
	return string(stdout)
}
