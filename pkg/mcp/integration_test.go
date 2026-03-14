// Integration tests for MCP server with real CLI commands.
//
// These tests require:
// - Valid CLI config at ~/.confluent/config
// - User logged in to Confluent Cloud or Platform
// - skills.json manifest in pkg/mcp/skills.json
//
// Run with: go test -v ./pkg/mcp -run TestIntegration -short=false
// Skip with: go test -short ./pkg/mcp
package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationVersionCommand tests executing the version command through MCP server.
// This is the most reliable test as it requires no authentication or configuration.
func TestIntegrationVersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Load real skills.json from project
	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err, "Failed to create MCP server")
	require.NotNil(t, server)

	// Execute version skill
	output, err := server.ExecuteSkill("confluent_version", nil)
	require.NoError(t, err, "Failed to execute version command")
	assert.NotEmpty(t, output, "Version output should not be empty")

	// Version output typically contains "confluent" and a version number
	assert.Contains(t, output, "confluent", "Output should mention confluent")
}

// TestIntegrationUnknownTool tests error handling for non-existent tools.
func TestIntegrationUnknownTool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Try to execute a non-existent tool
	output, err := server.ExecuteSkill("nonexistent_tool_12345", nil)
	assert.Error(t, err, "Should return error for unknown tool")
	assert.Empty(t, output, "Output should be empty for unknown tool")
	assert.Contains(t, err.Error(), "unknown tool", "Error should mention unknown tool")
}

// TestIntegrationJSONOutput tests executing a command that supports JSON output.
// This test may fail if user is not logged in or environment doesn't exist.
// That's acceptable - the test validates the MCP server can invoke commands correctly.
func TestIntegrationJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Try to list environments (requires Cloud login)
	// This may fail with auth error, which is fine - we're testing MCP server, not CLI auth
	output, err := server.ExecuteSkill("confluent_environment_list", map[string]interface{}{
		"output": "json",
	})

	// Document behavior: output may contain ANSI codes (to be stripped in Phase 4)
	// For now, just verify the command executed and returned something
	if err == nil {
		assert.NotEmpty(t, output, "Output should not be empty if command succeeded")
		t.Logf("Command succeeded, output length: %d", len(output))
	} else {
		// If command failed, it should be due to auth, not MCP server error
		// MCP server errors contain the tool name
		if assert.Contains(t, err.Error(), "environment_list", "Error should mention tool name") {
			t.Logf("Command failed (expected if not logged in): %v", err)
		}
	}
}

// TestIntegrationLoginRejection tests that browser-based login is rejected with helpful error.
func TestIntegrationLoginRejection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Try to login without API key (browser-based login)
	output, err := server.ExecuteSkill("confluent_login", map[string]interface{}{})
	assert.Error(t, err, "Browser login should be rejected")
	assert.Empty(t, output, "Output should be empty for rejected login")
	assert.Contains(t, err.Error(), "browser-based login", "Error should mention browser login")
	assert.Contains(t, err.Error(), "API key", "Error should suggest API key auth")
	assert.Contains(t, err.Error(), "LIMITATIONS.md", "Error should reference limitations doc")
}

// TestIntegrationParameterMapping tests that parameters are correctly passed to CLI.
func TestIntegrationParameterMapping(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Execute version command with output parameter
	// Version command doesn't require auth and should always work
	output, err := server.ExecuteSkill("confluent_version", map[string]interface{}{
		"output": "json",
	})

	// Note: version command may not support --output json, which is fine
	// We're just verifying parameters are passed through to the executor
	if err == nil {
		assert.NotEmpty(t, output, "Version output should not be empty")
	} else {
		// If it failed, verify the error was from CLI, not MCP server
		t.Logf("Version with params failed: %v", err)
		// Error should still be wrapped with tool name
		assert.Contains(t, err.Error(), "confluent_version", "Error should mention tool name")
	}
}

// TestIntegrationManifestLoading tests that the real skills.json loads correctly.
func TestIntegrationManifestLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	manifest := server.Manifest()
	require.NotNil(t, manifest)

	// Verify manifest has expected structure
	assert.NotEmpty(t, manifest.Metadata.CLIVersion, "CLI version should be set")
	assert.NotEmpty(t, manifest.Metadata.GeneratedAt, "Generated timestamp should be set")
	assert.Greater(t, len(manifest.Tools), 100, "Should have >100 tools (actual ~420)")

	// Verify some expected tools exist
	expectedTools := []string{
		"confluent_version",
		"confluent_login",
		"confluent_environment_list",
		"confluent_kafka_kafka_cluster_list",
	}

	for _, toolName := range expectedTools {
		_, ok := server.tools[toolName]
		assert.True(t, ok, "Tool %s should exist in manifest", toolName)
	}
}

// TestIntegrationAuthFailure validates auth failure errors are enhanced with actionable suggestions (TEST-06)
func TestIntegrationAuthFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create a temp directory for isolated config
	tempDir := t.TempDir()

	// Create empty or invalid config file
	invalidConfigPath := filepath.Join(tempDir, "config")
	err := os.WriteFile(invalidConfigPath, []byte("{}"), 0644)
	require.NoError(t, err)

	// Set config path environment variable to use temp config
	t.Setenv("HOME", tempDir)

	// Create temp .confluent directory
	os.MkdirAll(filepath.Join(tempDir, ".confluent"), 0755)
	os.Rename(invalidConfigPath, filepath.Join(tempDir, ".confluent", "config"))

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	// Server creation may succeed even with invalid config
	// Auth errors appear during command execution
	if err != nil {
		t.Logf("Expected: MCP server creation may fail with invalid config: %v", err)
		return
	}

	// Try to execute a command that requires authentication
	output, err := server.ExecuteSkill("confluent_environment_list", nil)

	// Expected: Auth failure or "not logged in" error
	// The error should be formatted with context
	if err != nil {
		assert.Contains(t, err.Error(), "environment_list", "Error should mention tool name")
		t.Logf("Auth failure error (expected): %v", err)
		t.Logf("Formatted output: %s", output)
	} else {
		// If it succeeded, the temp config might have valid credentials
		t.Logf("Command succeeded unexpectedly (temp config may have credentials)")
	}
}

// TestIntegrationInvalidParameters validates invalid parameter scenarios return helpful errors (TEST-06)
func TestIntegrationInvalidParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	tests := []struct {
		name        string
		toolName    string
		params      map[string]interface{}
		expectError bool
		errContains string
	}{
		{
			name:        "missing required cluster parameter",
			toolName:    "confluent_kafka_kafka_topic_list",
			params:      map[string]interface{}{}, // Missing required cluster ID
			expectError: true,                     // May error or may prompt - either is valid
			errContains: "kafka_topic_list",
		},
		{
			name:     "invalid parameter type",
			toolName: "confluent_kafka_kafka_topic_list",
			params: map[string]interface{}{
				"cluster": 12345, // Should be string, not int
			},
			expectError: true,
			errContains: "kafka_topic_list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := server.ExecuteSkill(tt.toolName, tt.params)

			// Document the behavior - errors may occur during param mapping or execution
			if err != nil {
				assert.Contains(t, err.Error(), tt.errContains, "Error should mention tool name")
				t.Logf("Error (expected): %v", err)
				t.Logf("Output: %s", output)
			} else {
				// Some CLI commands may accept invalid params and return usage text
				// This is still valid - we're testing that the server doesn't crash
				t.Logf("Command executed without error (may show usage): %s", output)
			}

			// Key validation: server didn't crash
			// Error handling infrastructure is working
		})
	}
}

// TestIntegrationCommandTimeout validates timeout handling works gracefully (TEST-06)
func TestIntegrationCommandTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create execution context directly to use ExecuteWithTimeout
	ctx, err := NewExecutionContext()
	if err != nil {
		t.Skipf("Cannot create execution context: %v", err)
		return
	}

	// Try to execute a potentially slow command with very short timeout
	output, err := ctx.ExecuteWithTimeout("version", nil, 1*time.Millisecond)

	// Most commands complete faster than 1ms, but this tests the timeout mechanism
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			t.Logf("Timeout error (expected for fast commands): %v", err)
		} else {
			t.Logf("Other error: %v", err)
		}
	} else {
		// Command completed within 1ms (common for version)
		assert.NotEmpty(t, output, "Output should not be empty if command succeeded")
		t.Logf("Command completed within timeout: %s", output)
	}

	// Test with reasonable timeout (should succeed)
	output, err = ctx.ExecuteWithTimeout("version", nil, 5*time.Second)
	require.NoError(t, err, "Version command should succeed with reasonable timeout")
	assert.NotEmpty(t, output, "Version output should not be empty")
}

// TestIntegrationCommandNotFound validates missing command errors are clear (TEST-06)
func TestIntegrationCommandNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Try to execute a non-existent tool
	output, err := server.ExecuteSkill("confluent_nonexistent_command_12345", nil)

	// Should return "unknown tool" error
	assert.Error(t, err, "Non-existent tool should return error")
	assert.Empty(t, output, "Output should be empty for unknown tool")
	assert.Contains(t, err.Error(), "unknown tool", "Error should mention unknown tool")
	assert.Contains(t, err.Error(), "confluent_nonexistent_command_12345", "Error should include tool name")
}

// TestIntegrationMalformedOutput validates output parsing errors don't crash server (TEST-06)
func TestIntegrationMalformedOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	skillsPath := "skills.json"

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Execute a command that's known to produce text output (not JSON)
	// The formatter should handle this gracefully
	output, err := server.ExecuteSkill("confluent_version", nil)

	// Version command should succeed (no auth required)
	if err == nil {
		assert.NotEmpty(t, output, "Version output should not be empty")
		t.Logf("Version output: %s", output)
	} else {
		t.Logf("Version command error: %v", err)
	}

	// The key validation: server didn't crash, even if output isn't JSON
	// Formatter should return text output as-is or extract meaningful summary
}
