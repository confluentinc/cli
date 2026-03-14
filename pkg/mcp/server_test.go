package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/confluentinc/cli/v4/pkg/skillgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutionContext implements ExecutionContext interface for testing
type mockExecutionContext struct {
	executeCalls []mockExecuteCall
	executeFunc  func(commandPath string, params map[string]interface{}) (string, error)
}

type mockExecuteCall struct {
	commandPath string
	params      map[string]interface{}
}

func (m *mockExecutionContext) Execute(commandPath string, params map[string]interface{}) (string, error) {
	m.executeCalls = append(m.executeCalls, mockExecuteCall{
		commandPath: commandPath,
		params:      params,
	})
	if m.executeFunc != nil {
		return m.executeFunc(commandPath, params)
	}
	return "mock output", nil
}

func (m *mockExecutionContext) ExecuteWithTimeout(commandPath string, params map[string]interface{}, timeout time.Duration) (string, error) {
	return m.Execute(commandPath, params)
}

// createTestSkillsJSON creates a minimal test skills.json file
func createTestSkillsJSON(t *testing.T) string {
	tmpDir := t.TempDir()
	skillsPath := filepath.Join(tmpDir, "skills.json")

	manifest := skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "4.0.0",
			GeneratedAt:      "2026-03-10T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       3,
			CommandCount:     3,
		},
		Tools: []skillgen.Tool{
			{
				Name:        "confluent_version",
				Title:       "Display Confluent CLI version",
				Description: "Display the version of the Confluent CLI.",
				CommandPath: "version",
				InputSchema: skillgen.InputSchema{
					Type:       "object",
					Properties: map[string]skillgen.JSONSchema{},
					Required:   []string{},
				},
				Annotations: skillgen.Annotations{
					Audience: "customers",
					Priority: "high",
				},
			},
			{
				Name:        "confluent_kafka_cluster_list",
				Title:       "List Kafka clusters",
				Description: "List all Kafka clusters.",
				CommandPath: "kafka cluster list",
				InputSchema: skillgen.InputSchema{
					Type: "object",
					Properties: map[string]skillgen.JSONSchema{
						"environment": {
							Type:        "string",
							Description: "Environment ID",
						},
					},
					Required: []string{},
				},
				Annotations: skillgen.Annotations{
					Audience: "customers",
					Priority: "high",
				},
			},
			{
				Name:        "confluent_login",
				Title:       "Login to Confluent",
				Description: "Login to Confluent Cloud or Platform.",
				CommandPath: "login",
				InputSchema: skillgen.InputSchema{
					Type: "object",
					Properties: map[string]skillgen.JSONSchema{
						"url": {
							Type:        "string",
							Description: "Platform URL",
						},
					},
					Required: []string{},
				},
				Annotations: skillgen.Annotations{
					Audience: "customers",
					Priority: "high",
				},
			},
		},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(skillsPath, data, 0644))

	return skillsPath
}

func TestNewMCPServerLoadsManifest(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)
	require.NotNil(t, server)

	assert.NotNil(t, server.Manifest())
	assert.Equal(t, 3, len(server.Manifest().Tools))
	assert.Equal(t, "4.0.0", server.Manifest().Metadata.CLIVersion)
}

func TestNewMCPServerManifestError(t *testing.T) {
	server, err := NewMCPServer("/nonexistent/skills.json")
	assert.Error(t, err)
	assert.Nil(t, server)
	assert.Contains(t, err.Error(), "skills.json")
}

func TestNewMCPServerBuildToolMap(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Verify tool map was built
	assert.Equal(t, 3, len(server.tools))
	assert.NotNil(t, server.tools["confluent_version"])
	assert.NotNil(t, server.tools["confluent_kafka_cluster_list"])
	assert.NotNil(t, server.tools["confluent_login"])
}

func TestExecuteSkillUnknownTool(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	output, err := server.ExecuteSkill("nonexistent_tool", nil)
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "unknown tool")
}

func TestExecuteSkillCallsExecute(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Replace execution context with mock
	mockCtx := &mockExecutionContext{
		executeFunc: func(commandPath string, params map[string]interface{}) (string, error) {
			return "version output", nil
		},
	}
	server.ctx = mockCtx

	output, err := server.ExecuteSkill("confluent_version", nil)
	require.NoError(t, err)
	// Output should be formatted (ANSI stripped, plain text passed through)
	assert.Equal(t, "version output", output)

	// Verify Execute was called
	require.Equal(t, 1, len(mockCtx.executeCalls))
	assert.Equal(t, "version", mockCtx.executeCalls[0].commandPath)
}

func TestExecuteSkillWrapsErrors(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Replace execution context with mock that returns error
	mockCtx := &mockExecutionContext{
		executeFunc: func(commandPath string, params map[string]interface{}) (string, error) {
			return "", assert.AnError
		},
	}
	server.ctx = mockCtx

	output, err := server.ExecuteSkill("confluent_version", nil)
	assert.Error(t, err)
	// Output now contains formatted error message (not empty)
	assert.Contains(t, output, "Command: version")
	assert.Contains(t, output, "Error:")
	assert.Contains(t, err.Error(), "confluent_version")
}

func TestExecuteSkillDetectsBrowserLogin(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Execute login without API key params - should be rejected
	output, err := server.ExecuteSkill("confluent_login", map[string]interface{}{
		"url": "https://platform.example.com",
	})
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "browser-based login")
	assert.Contains(t, err.Error(), "API key")
}

func TestExecuteSkillAllowsAPIKeyLogin(t *testing.T) {
	skillsPath := createTestSkillsJSON(t)

	server, err := NewMCPServer(skillsPath)
	require.NoError(t, err)

	// Replace execution context with mock
	mockCtx := &mockExecutionContext{
		executeFunc: func(commandPath string, params map[string]interface{}) (string, error) {
			return "login success", nil
		},
	}
	server.ctx = mockCtx

	// Execute login WITH API key params - should be allowed
	output, err := server.ExecuteSkill("confluent_login", map[string]interface{}{
		"api-key":    "APIKEY123",
		"api-secret": "secret",
	})
	require.NoError(t, err)
	// Output should be formatted
	assert.Equal(t, "login success", output)
}

func TestExecuteSkillFormatsOutput(t *testing.T) {
	t.Run("strips ANSI codes", func(t *testing.T) {
		skillsPath := createTestSkillsJSON(t)
		server, err := NewMCPServer(skillsPath)
		require.NoError(t, err)

		mockCtx := &mockExecutionContext{
			executeFunc: func(commandPath string, params map[string]interface{}) (string, error) {
				return "\x1b[32mGreen text\x1b[0m", nil
			},
		}
		server.ctx = mockCtx

		output, err := server.ExecuteSkill("confluent_version", nil)
		require.NoError(t, err)
		assert.Equal(t, "Green text", output)
		assert.NotContains(t, output, "\x1b")
	})

	t.Run("formats JSON arrays as markdown tables", func(t *testing.T) {
		skillsPath := createTestSkillsJSON(t)
		server, err := NewMCPServer(skillsPath)
		require.NoError(t, err)

		mockCtx := &mockExecutionContext{
			executeFunc: func(commandPath string, params map[string]interface{}) (string, error) {
				return `[{"id":"lkc-1","name":"prod"},{"id":"lkc-2","name":"dev"}]`, nil
			},
		}
		server.ctx = mockCtx

		output, err := server.ExecuteSkill("confluent_kafka_cluster_list", nil)
		require.NoError(t, err)
		assert.Contains(t, output, "Found 2 items:")
		assert.Contains(t, output, "| Id | Name |")
		assert.Contains(t, output, "| lkc-1 | prod |")
	})

	t.Run("formats errors with command context", func(t *testing.T) {
		skillsPath := createTestSkillsJSON(t)
		server, err := NewMCPServer(skillsPath)
		require.NoError(t, err)

		mockCtx := &mockExecutionContext{
			executeFunc: func(commandPath string, params map[string]interface{}) (string, error) {
				return "", assert.AnError
			},
		}
		server.ctx = mockCtx

		output, err := server.ExecuteSkill("confluent_version", nil)
		assert.Error(t, err)
		assert.Contains(t, output, "Command: version")
		assert.Contains(t, output, "Error:")
	})
}

func TestHasAPIKeyParam(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		want   bool
	}{
		{
			name:   "no params",
			params: nil,
			want:   false,
		},
		{
			name:   "empty params",
			params: map[string]interface{}{},
			want:   false,
		},
		{
			name: "has api-key",
			params: map[string]interface{}{
				"api-key": "APIKEY123",
			},
			want: true,
		},
		{
			name: "has api-secret",
			params: map[string]interface{}{
				"api-secret": "secret",
			},
			want: true,
		},
		{
			name: "has both",
			params: map[string]interface{}{
				"api-key":    "APIKEY123",
				"api-secret": "secret",
			},
			want: true,
		},
		{
			name: "other params only",
			params: map[string]interface{}{
				"url": "https://example.com",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAPIKeyParam(tt.params)
			assert.Equal(t, tt.want, got)
		})
	}
}
