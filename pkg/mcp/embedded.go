package mcp

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/confluentinc/cli/v4/pkg/skillgen"
)

//go:embed skills.json
var embeddedSkillsManifest []byte

// LoadEmbeddedSkills loads the skills manifest from embedded data.
// Returns an error if the manifest is not embedded (development builds).
func LoadEmbeddedSkills() (*skillgen.SkillManifest, error) {
	if len(embeddedSkillsManifest) == 0 {
		return nil, fmt.Errorf("skills manifest not embedded in binary (development build)")
	}

	var manifest skillgen.SkillManifest
	if err := json.Unmarshal(embeddedSkillsManifest, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshaling embedded skills manifest: %w", err)
	}

	return &manifest, nil
}

// LoadSkills loads the skills manifest with fallback logic.
// First attempts to load from embedded data (production builds).
// Falls back to loading from file (development builds).
func LoadSkills() (*skillgen.SkillManifest, error) {
	// Try embedded first (production)
	manifest, err := LoadEmbeddedSkills()
	if err == nil {
		return manifest, nil
	}

	// Fall back to file (development)
	return loadSkillsFromFile("pkg/mcp/skills.json")
}

// loadSkillsFromFile loads the skills manifest from a file path.
// This is used in development builds when the manifest is not embedded.
func loadSkillsFromFile(path string) (*skillgen.SkillManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading skills manifest from %s: %w", path, err)
	}

	var manifest skillgen.SkillManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshaling skills manifest from %s: %w", path, err)
	}

	return &manifest, nil
}
