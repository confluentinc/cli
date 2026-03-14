package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/cli/v4/pkg/mcp/formatter"
	"github.com/confluentinc/cli/v4/pkg/skillgen"
)

// Executor is the interface for executing CLI commands.
// This allows mocking the ExecutionContext in tests.
type Executor interface {
	Execute(commandPath string, params map[string]interface{}) (string, error)
	ExecuteWithTimeout(commandPath string, params map[string]interface{}, timeout time.Duration) (string, error)
}

// MCPServer manages skill loading and execution for the MCP protocol.
// It loads the skill manifest, builds a tool lookup map, and delegates
// execution to the ExecutionContext while handling browser login detection.
type MCPServer struct {
	ctx      Executor
	manifest *skillgen.SkillManifest
	tools    map[string]*skillgen.Tool
}

// NewMCPServer creates a new MCP server by loading the skills manifest
// from the specified path. It initializes the execution context and builds
// the tool lookup map for fast skill resolution.
//
// Returns an error if:
// - The skills.json file cannot be read
// - The JSON is malformed
// - The execution context cannot be initialized
func NewMCPServer(skillsPath string) (*MCPServer, error) {
	// Create execution context with CLI config
	ctx, err := NewExecutionContext()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize execution context: %w", err)
	}

	// Read skills manifest
	data, err := os.ReadFile(skillsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills.json: %w", err)
	}

	// Unmarshal manifest
	var manifest skillgen.SkillManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse skills.json: %w", err)
	}

	// Build tool lookup map
	tools := make(map[string]*skillgen.Tool)
	for i := range manifest.Tools {
		tool := &manifest.Tools[i]
		tools[tool.Name] = tool
	}

	return &MCPServer{
		ctx:      ctx,
		manifest: &manifest,
		tools:    tools,
	}, nil
}

// Manifest returns the loaded skill manifest for inspection.
// Useful for displaying available skills or debugging.
func (s *MCPServer) Manifest() *skillgen.SkillManifest {
	return s.manifest
}

// ExecuteSkill executes a skill by name with the given parameters.
// It looks up the tool, extracts the command path, detects browser login
// attempts, and delegates to the ExecutionContext for actual execution.
// The output is formatted via the formatter before being returned.
//
// Returns an error if:
// - The tool name is not found
// - Browser-based login is attempted (EXEC-08 limitation)
// - The command execution fails
func (s *MCPServer) ExecuteSkill(toolName string, params map[string]interface{}) (string, error) {
	// Look up tool by name
	tool, ok := s.tools[toolName]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	// Use the command path stored in the tool manifest
	commandPath := tool.CommandPath

	// Detect browser-based login (EXEC-08 limitation)
	if isBrowserLogin(commandPath, params) {
		return "", fmt.Errorf("browser-based login cannot be invoked from MCP server. Use API key authentication instead. See pkg/mcp/LIMITATIONS.md for details")
	}

	// Execute command via ExecutionContext
	output, err := s.ctx.Execute(commandPath, params)

	// Format output before returning
	f := formatter.NewFormatter()
	formatted := f.Format(output, err, commandPath)

	// Return formatted summary
	if err != nil {
		return formatted.Summary, fmt.Errorf("executing %s: %w", toolName, err)
	}
	return formatted.Summary, nil
}

// isBrowserLogin detects if a command is a browser-based login attempt.
// Browser login is incompatible with MCP server (EXEC-08 limitation).
//
// Detection logic:
// - Command path contains "login"
// - Parameters do NOT contain "api-key" or "api-secret" (API key auth is OK)
//
// Returns true if this is a browser login that should be rejected.
func isBrowserLogin(commandPath string, params map[string]interface{}) bool {
	// Only login commands are affected
	if !strings.Contains(commandPath, "login") {
		return false
	}

	// API key login is allowed (non-interactive)
	if hasAPIKeyParam(params) {
		return false
	}

	// Browser-based login detected
	return true
}

// hasAPIKeyParam checks if parameters contain API key authentication.
// Used to distinguish interactive browser login from API key login.
func hasAPIKeyParam(params map[string]interface{}) bool {
	if params == nil {
		return false
	}

	// Check for api-key or api-secret parameters
	_, hasAPIKey := params["api-key"]
	_, hasAPISecret := params["api-secret"]

	return hasAPIKey || hasAPISecret
}
