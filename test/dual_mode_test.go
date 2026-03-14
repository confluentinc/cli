//go:build integration
// +build integration

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/confluentinc/cli/v4/pkg/mcp"
	"github.com/confluentinc/cli/v4/pkg/skillgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTempConfluent creates a temporary CONFLUENT_HOME for test isolation
func setupTempConfluent(t *testing.T) string {
	tempDir := t.TempDir()

	// Set environment variable for config location
	t.Setenv("CONFLUENT_HOME", tempDir)

	return tempDir
}

// createDualModeTestManifest creates a minimal skills manifest for testing
func createDualModeTestManifest(t *testing.T) string {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "skills.json")

	manifest := skillgen.SkillManifest{
		Metadata: skillgen.ManifestMetadata{
			CLIVersion:       "4.0.0",
			GeneratedAt:      "2026-03-13T00:00:00Z",
			GeneratorVersion: "1.0.0",
			SkillCount:       2,
			CommandCount:     2,
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
				Description: "List all Kafka clusters in an environment.",
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
		},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	return manifestPath
}

func TestDualMode_ManifestLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping dual-mode integration test in short mode")
	}

	// Setup isolated temp directory
	setupTempConfluent(t)

	// Create test manifest
	manifestPath := createDualModeTestManifest(t)

	// Load MCP server (initializes execution context)
	server, err := mcp.NewMCPServer(manifestPath)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Verify manifest loaded with correct tool count
	manifest := server.Manifest()
	require.NotNil(t, manifest)
	assert.Equal(t, 2, len(manifest.Tools))

	// Verify tool lookup works for both tools
	foundVersion := false
	foundKafka := false
	for _, tool := range manifest.Tools {
		switch tool.Name {
		case "confluent_version":
			foundVersion = true
		case "confluent_kafka_cluster_list":
			foundKafka = true
		}
	}
	assert.True(t, foundVersion, "confluent_version tool should be available")
	assert.True(t, foundKafka, "confluent_kafka_cluster_list tool should be available")

	// Verify unknown tool returns error
	_, err = server.ExecuteSkill("nonexistent_tool", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}
