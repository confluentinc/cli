package mcp

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/internal"
	"github.com/confluentinc/cli/v4/pkg/config"
)

// outputMutex serializes all command execution because output capture requires
// redirecting the global os.Stdout/os.Stderr. This means concurrent skill
// invocations are executed sequentially. This is a known architectural
// limitation of in-process output capture; eliminating it would require all
// CLI commands to use cmd.OutOrStdout() instead of writing to os.Stdout directly.
var outputMutex sync.Mutex

// Execute runs a CLI command in-process with output capture.
// Returns the combined stdout+stderr output and any error.
//
// Parameters:
//   - cfg: CLI configuration (preserves authentication context)
//   - commandPath: Space-separated command path (e.g., "kafka cluster list")
//   - params: Map of parameter names to values (will be mapped to CLI flags)
//
// The function:
//  1. Builds a fresh command tree from the config
//  2. Parses commandPath into args and finds the target subcommand
//  3. Maps parameters to CLI flags via MapParamsToFlags
//  4. Forces JSON output via ForceJSONOutput (if --output flag exists)
//  5. Captures stdout/stderr during execution with mutex protection
//  6. Returns combined output and any error
func Execute(cfg *config.Config, commandPath string, params map[string]interface{}) (string, error) {
	// Build fresh command tree
	cmd := internal.NewConfluentCommand(cfg)

	// Parse command path into args
	args := strings.Fields(commandPath)
	if len(args) == 0 {
		return "", fmt.Errorf("empty command path")
	}

	// Find target subcommand
	targetCmd, _, err := cmd.Find(args)
	if err != nil {
		return "", err
	}

	// Map parameters to flags (if provided)
	if params != nil {
		if err := MapParamsToFlags(targetCmd, params); err != nil {
			return "", err
		}
	}

	// Force JSON output when supported
	if err := ForceJSONOutput(targetCmd); err != nil {
		return "", err
	}

	// Capture and execute
	return captureExecute(targetCmd)
}

// captureExecute executes a command while capturing its stdout and stderr output.
// It uses a mutex to ensure thread-safe output redirection and restores the
// original stdout/stderr after execution.
//
// The pipe is read concurrently with command execution to prevent deadlock
// when command output exceeds the OS pipe buffer (~64KB).
//
// Returns the combined output (stdout + stderr) and any error from cmd.Execute().
func captureExecute(cmd *cobra.Command) (string, error) {
	// Lock mutex for thread-safe stdout/stderr manipulation
	outputMutex.Lock()
	defer outputMutex.Unlock()

	// Save original stdout and stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	defer func() {
		// Restore original stdout/stderr
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	// Redirect stdout and stderr to pipe
	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("failed to create pipe: %w", err)
	}

	os.Stdout = w
	os.Stderr = w

	// Read from pipe concurrently to prevent deadlock when
	// command output exceeds the OS pipe buffer (~64KB).
	var stdoutBuf bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		stdoutBuf.ReadFrom(r)
		close(readDone)
	}()

	// Execute command
	execErr := cmd.Execute()

	// Close write end, wait for reader to drain, then close read end
	w.Close()
	<-readDone
	r.Close()

	return stdoutBuf.String(), execErr
}
