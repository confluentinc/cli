package mcp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadEmbeddedSkills tests loading the manifest from embedded data.
// This test will fail in development builds where skills.json is not
// properly embedded (empty byte slice), which is expected behavior.
func TestLoadEmbeddedSkills(t *testing.T) {
	manifest, err := LoadEmbeddedSkills()

	// In development builds, embedded data may be empty (placeholder file)
	if len(embeddedSkillsManifest) == 0 {
		require.Error(t, err)
		require.Contains(t, err.Error(), "not embedded")
		return
	}

	// In production builds (or with real data), should succeed
	require.NoError(t, err)
	require.NotNil(t, manifest)
	require.NotNil(t, manifest.Metadata)
	require.NotEmpty(t, manifest.Metadata.CLIVersion)
}

// TestLoadSkillsFallback tests that LoadSkills falls back to file loading
// when embedded data is unavailable.
func TestLoadSkillsFallback(t *testing.T) {
	// Create a temporary skills.json file
	tmpDir := t.TempDir()
	skillsPath := tmpDir + "/skills.json"

	// Write test manifest to file
	data := `{
  "metadata": {
    "cli_version": "v3.45.0",
    "generated_at": "2026-03-12T00:00:00Z",
    "generator_version": "1.0.0",
    "skill_count": 2,
    "command_count": 5
  },
  "tools": [
    {
      "name": "test-tool",
      "title": "Test Tool",
      "description": "A test tool",
      "inputSchema": {
        "type": "object",
        "properties": {},
        "required": []
      },
      "annotations": {}
    }
  ]
}`
	require.NoError(t, os.WriteFile(skillsPath, []byte(data), 0644))

	// Test loadSkillsFromFile directly
	manifest, err := loadSkillsFromFile(skillsPath)
	require.NoError(t, err)
	require.NotNil(t, manifest)
	require.Equal(t, "v3.45.0", manifest.Metadata.CLIVersion)
	require.Equal(t, 2, manifest.Metadata.SkillCount)
	require.Len(t, manifest.Tools, 1)
	require.Equal(t, "test-tool", manifest.Tools[0].Name)
}

// TestLoadSkillsFromFileError tests error handling when file doesn't exist.
func TestLoadSkillsFromFileError(t *testing.T) {
	_, err := loadSkillsFromFile("/nonexistent/path/skills.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading skills manifest")
}

