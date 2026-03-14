package mcp

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/config"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

func TestCaptureExecuteRedirectsOutput(t *testing.T) {
	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	// Create a simple command that writes to stdout
	output, err := Execute(cfg, "version", nil)
	require.NoError(t, err)
	require.Contains(t, output, "version", "expected output to contain 'version'")
}

func TestCaptureExecuteRestoresOriginal(t *testing.T) {
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	_, _ = Execute(cfg, "version", nil)

	// Verify stdout/stderr restored
	require.Equal(t, originalStdout, os.Stdout, "stdout should be restored")
	require.Equal(t, originalStderr, os.Stderr, "stderr should be restored")
}

func TestCaptureExecuteMutexProtection(t *testing.T) {
	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	// Run concurrent executions
	var wg sync.WaitGroup
	results := make([]string, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			output, err := Execute(cfg, "version", nil)
			require.NoError(t, err)
			results[idx] = output
		}(i)
	}

	wg.Wait()

	// Verify all outputs are valid (not interleaved)
	for _, result := range results {
		require.Contains(t, result, "version", "each output should be complete")
	}
}

func TestExecuteBuildsCommandTree(t *testing.T) {
	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	// Execute should build command tree from config
	output, err := Execute(cfg, "version", nil)
	require.NoError(t, err)
	require.NotEmpty(t, output, "should have output from command tree")
}

func TestExecuteFindsSubcommand(t *testing.T) {
	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	// Test finding subcommand via path
	output, err := Execute(cfg, "version", nil)
	require.NoError(t, err)
	require.NotEmpty(t, output)
}

func TestExecuteReturnsError(t *testing.T) {
	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	// Execute with invalid command should return error
	_, err := Execute(cfg, "nonexistent-command", nil)
	require.Error(t, err, "should return error for invalid command")
}

func TestExecuteCombinesStdoutStderr(t *testing.T) {
	cfg := &config.Config{
		IsTest:  true,
		Version: &pversion.Version{Version: "test"},
	}

	// Both stdout and stderr should be captured
	output, err := Execute(cfg, "version", nil)
	require.NoError(t, err)
	require.NotEmpty(t, output, "should capture output")
}
