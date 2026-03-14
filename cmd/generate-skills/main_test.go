package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/skillgen"
)

func TestMainProducesValidJSON(t *testing.T) {
	// Test that main() creates a valid ir.json file

	outputPath := "test-ir.json"
	defer os.Remove(outputPath)

	// Run main logic (will implement as testable function)
	err := generateIR(outputPath)
	require.NoError(t, err, "generateIR should not return error")

	// Verify file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err, "ir.json file should exist")

	// Read and parse JSON
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err, "should read generated file")

	var ir skillgen.IR
	err = json.Unmarshal(data, &ir)
	require.NoError(t, err, "generated JSON should be valid")

	// Verify structure
	assert.NotEmpty(t, ir.Commands, "IR should contain commands")
	assert.NotEmpty(t, ir.Metadata.CLIVersion, "IR should have CLI version")
	assert.False(t, ir.Metadata.GeneratedAt.IsZero(), "IR should have generation timestamp")
}

func TestJSONFormatting(t *testing.T) {
	// Test that JSON is pretty-printed with proper indentation

	outputPath := "test-ir-format.json"
	defer os.Remove(outputPath)

	err := generateIR(outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	// Check for indentation (pretty-printed JSON should have newlines and spaces)
	content := string(data)
	assert.Contains(t, content, "\n", "JSON should be formatted with newlines")
	assert.Contains(t, content, "  ", "JSON should have indentation")
	assert.Contains(t, content, "\"metadata\"", "JSON should contain metadata key")
	assert.Contains(t, content, "\"commands\"", "JSON should contain commands key")
}

func TestGenerateSkills(t *testing.T) {
	// Test that generateSkills creates valid skills.json file

	// First generate IR
	irPath := "test-ir-for-skills.json"
	skillsPath := "test-skills.json"
	defer os.Remove(irPath)
	defer os.Remove(skillsPath)

	err := generateIR(irPath)
	require.NoError(t, err, "generateIR should not return error")

	// Generate skills from IR
	err = generateSkills(irPath, skillsPath)
	require.NoError(t, err, "generateSkills should not return error")

	// Verify file exists
	_, err = os.Stat(skillsPath)
	require.NoError(t, err, "skills.json file should exist")

	// Read and parse JSON
	data, err := os.ReadFile(skillsPath)
	require.NoError(t, err, "should read generated skills file")

	var manifest skillgen.SkillManifest
	err = json.Unmarshal(data, &manifest)
	require.NoError(t, err, "generated JSON should be valid")

	// Verify structure
	assert.NotEmpty(t, manifest.Tools, "manifest should contain tools")
	assert.NotEmpty(t, manifest.Metadata.CLIVersion, "manifest should have CLI version")
	assert.NotEmpty(t, manifest.Metadata.GeneratedAt, "manifest should have generation timestamp")
	assert.NotEmpty(t, manifest.Metadata.GeneratorVersion, "manifest should have generator version")
	assert.Greater(t, manifest.Metadata.SkillCount, 0, "skill count should be positive")
	assert.Greater(t, manifest.Metadata.CommandCount, 0, "command count should be positive")
	assert.Equal(t, len(manifest.Tools), manifest.Metadata.SkillCount, "tool count should match metadata skill_count")
}

func TestGenerateSkillsWithInvalidIR(t *testing.T) {
	// Test that generateSkills returns error when IR file doesn't exist

	skillsPath := "test-skills-invalid.json"
	defer os.Remove(skillsPath)

	err := generateSkills("nonexistent-ir.json", skillsPath)
	require.Error(t, err, "generateSkills should return error for missing IR")
	assert.Contains(t, err.Error(), "reading IR", "error should mention reading IR")
}

func TestSkillsJSONFormatting(t *testing.T) {
	// Test that skills.json is pretty-printed with proper indentation

	irPath := "test-ir-for-formatting.json"
	skillsPath := "test-skills-format.json"
	defer os.Remove(irPath)
	defer os.Remove(skillsPath)

	err := generateIR(irPath)
	require.NoError(t, err)

	err = generateSkills(irPath, skillsPath)
	require.NoError(t, err)

	data, err := os.ReadFile(skillsPath)
	require.NoError(t, err)

	// Check for indentation (2-space indent, consistent with ir.json)
	content := string(data)
	assert.Contains(t, content, "\n", "JSON should be formatted with newlines")
	assert.Contains(t, content, "  ", "JSON should have 2-space indentation")
	assert.Contains(t, content, "\"metadata\"", "JSON should contain metadata key")
	assert.Contains(t, content, "\"tools\"", "JSON should contain tools key")
}
