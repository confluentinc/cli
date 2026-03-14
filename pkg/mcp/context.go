package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/cli/v4/pkg/config"
)

// ExecutionContext manages CLI config and provides timeout-capable command execution.
// It loads the CLI config once and reuses it across all skill invocations to preserve
// authentication state.
type ExecutionContext struct {
	cfg *config.Config
	mu  sync.Mutex
}

// NewExecutionContext creates a new execution context by loading the CLI config.
// Returns an error if the config file cannot be loaded or is invalid.
func NewExecutionContext() (*ExecutionContext, error) {
	// Load config from default location (~/.confluent/config)
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &ExecutionContext{
		cfg: cfg,
	}, nil
}

// Config returns the loaded CLI configuration.
// The config is immutable after construction so no synchronization is needed.
func (ec *ExecutionContext) Config() *config.Config {
	return ec.cfg
}

// Execute runs a CLI command with the given parameters.
// The config is passed to the executor to preserve authentication context.
func (ec *ExecutionContext) Execute(commandPath string, params map[string]interface{}) (string, error) {
	ec.mu.Lock()
	cfg := ec.cfg
	ec.mu.Unlock()

	// Call the executor with the config
	return Execute(cfg, commandPath, params)
}

// result holds the output and error from an async Execute call
type result struct {
	output string
	err    error
}

// ExecuteWithTimeout runs a CLI command with a timeout.
// If timeout is 0, no timeout is enforced (calls Execute directly).
// If the command exceeds the timeout, returns a timeout error.
//
// Note: the timeout only controls when the caller receives an error — the
// underlying command goroutine continues running to completion because
// in-process Cobra commands cannot be interrupted. The goroutine will not
// leak (the buffered channel ensures it can send its result) but it will
// continue holding the output mutex until the command finishes.
func (ec *ExecutionContext) ExecuteWithTimeout(commandPath string, params map[string]interface{}, timeout time.Duration) (string, error) {
	// If no timeout specified, execute directly
	if timeout == 0 {
		return ec.Execute(commandPath, params)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create buffered channel for result (prevents goroutine leak if timeout fires first)
	resultCh := make(chan result, 1)

	// Execute in goroutine
	go func() {
		output, err := ec.Execute(commandPath, params)
		resultCh <- result{output: output, err: err}
	}()

	// Wait for either result or timeout
	select {
	case res := <-resultCh:
		// Command completed before timeout
		return res.output, res.err
	case <-ctx.Done():
		// Timeout occurred
		return "", fmt.Errorf("command execution timeout after %v", timeout)
	}
}
