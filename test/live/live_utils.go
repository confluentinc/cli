//go:build live_test

package live

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// uniqueName generates a unique resource name with the given prefix.
func uniqueName(prefix string) string {
	return fmt.Sprintf("cli-live-%s-%06d", prefix, rand.Intn(1000000))
}

// liveTestCloud returns the cloud provider from env or defaults to "aws".
func liveTestCloud() string {
	if v := os.Getenv("CLI_LIVE_TEST_CLOUD"); v != "" {
		return v
	}
	return "aws"
}

// liveTestRegion returns the cloud region from env or defaults to "us-east-1".
func liveTestRegion() string {
	if v := os.Getenv("CLI_LIVE_TEST_REGION"); v != "" {
		return v
	}
	return "us-east-1"
}

// requiredEnv gets an environment variable or fails the test.
func requiredEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

// shellSplit splits a command string into arguments, respecting double and single quotes.
// Unlike strings.Fields, it treats quoted substrings as single arguments.
func shellSplit(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	var quoteChar byte

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote:
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		case c == '"' || c == '\'':
			inQuote = true
			quoteChar = c
		case c == ' ' || c == '\t':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// extractJSONField extracts a specific field value from JSON output.
func extractJSONField(t *testing.T, output, field string) string {
	t.Helper()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		return ""
	}
	if val, ok := parsed[field]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// waitForCondition polls a CLI command at the given interval until condition returns true or timeout expires.
// The command template supports {{.key}} substitution from state. Command failures during polling are
// tolerated (the output is still checked against condition). Returns the last output on success.
func (s *CLILiveTestSuite) waitForCondition(t *testing.T, argsTemplate string, state *LiveTestState,
	condition func(output string) bool, interval, timeout time.Duration) string {
	t.Helper()

	deadline := time.Now().Add(timeout)
	args := substituteStateVars(t, argsTemplate, state)
	env := buildCommandEnv([]string{homeEnvVar(state.homeDir)})

	var lastOutput string
	for {
		cmd := exec.Command(s.binPath, shellSplit(args)...)
		cmd.Env = env
		out, _ := cmd.CombinedOutput()
		lastOutput = string(out)

		if condition(lastOutput) {
			return lastOutput
		}
		if time.Now().After(deadline) {
			t.Fatalf("waitForCondition timed out after %s — last output:\n%s", timeout, lastOutput)
		}
		t.Logf("Condition not met, retrying in %s...", interval)
		time.Sleep(interval)
	}
}
