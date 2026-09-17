//go:build eval

package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// minimal mirror of the fields the grader needs; avoids importing the full config load path.
type evalConfig struct {
	CurrentContext string `json:"current_context"`
	Contexts       map[string]struct {
		CurrentEnvironment string `json:"current_environment"`
	} `json:"contexts"`
}

func configPath(homeDir string) string {
	return filepath.Join(homeDir, ".confluent", "config.json")
}

func loadConfig(homeDir string) (*evalConfig, error) {
	data, err := os.ReadFile(configPath(homeDir))
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c evalConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &c, nil
}

// ReadActiveEnvironment returns the current context's active environment id.
func ReadActiveEnvironment(homeDir string) (string, error) {
	c, err := loadConfig(homeDir)
	if err != nil {
		return "", err
	}
	ctx, ok := c.Contexts[c.CurrentContext]
	if !ok {
		return "", fmt.Errorf("current_context %q missing from contexts", c.CurrentContext)
	}
	return ctx.CurrentEnvironment, nil
}

// GradeConfigIntegrity fails if the config is missing, unparseable, or structurally incomplete.
func GradeConfigIntegrity(homeDir string) error {
	c, err := loadConfig(homeDir)
	if err != nil {
		return err
	}
	if c.CurrentContext == "" {
		return fmt.Errorf("config has empty current_context")
	}
	if _, ok := c.Contexts[c.CurrentContext]; !ok {
		return fmt.Errorf("current_context %q missing from contexts", c.CurrentContext)
	}
	return nil
}
