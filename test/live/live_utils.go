//go:build live_test

package live

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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

// liveTestClusterType returns the cluster type from env or defaults to "basic".
func liveTestClusterType() string {
	if v := os.Getenv("CLI_LIVE_TEST_CLUSTER_TYPE"); v != "" {
		return v
	}
	return "basic"
}

// CloudVariant represents a cloud provider, region, and cluster type combination for parameterized tests.
type CloudVariant struct {
	Cloud       string
	Region      string
	ClusterType string
}

// String returns a human-readable label for the variant (used as t.Run subtest name).
func (v CloudVariant) String() string {
	return fmt.Sprintf("%s/%s/%s", v.Cloud, v.Region, v.ClusterType)
}

// defaultRegions maps cloud providers to their default regions.
var defaultRegions = map[string]string{
	"aws":   "us-east-1",
	"gcp":   "us-east1",
	"azure": "eastus",
}

// liveTestVariants parses CLI_LIVE_TEST_VARIANTS into a list of CloudVariant.
// Format: "cloud:region:type,cloud:region:type,..."
// Falls back to a single variant from CLI_LIVE_TEST_CLOUD/CLI_LIVE_TEST_REGION/CLI_LIVE_TEST_CLUSTER_TYPE.
func liveTestVariants() []CloudVariant {
	if v := os.Getenv("CLI_LIVE_TEST_VARIANTS"); v != "" {
		var variants []CloudVariant
		for _, entry := range strings.Split(v, ",") {
			parts := strings.SplitN(strings.TrimSpace(entry), ":", 3)
			variant := CloudVariant{Cloud: parts[0]}
			if len(parts) > 1 && parts[1] != "" {
				variant.Region = parts[1]
			} else {
				variant.Region = defaultRegions[variant.Cloud]
			}
			if len(parts) > 2 && parts[2] != "" {
				variant.ClusterType = parts[2]
			} else {
				variant.ClusterType = "basic"
			}
			variants = append(variants, variant)
		}
		return variants
	}
	return []CloudVariant{{
		Cloud:       liveTestCloud(),
		Region:      liveTestRegion(),
		ClusterType: liveTestClusterType(),
	}}
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

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}
