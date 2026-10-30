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
	CurrentContext string                 `json:"current_context"`
	Contexts       map[string]evalContext `json:"contexts"`
	ContextStates  map[string]struct {
		AuthToken string `json:"auth_token"`
	} `json:"context_states"`
}

type evalContext struct {
	CurrentEnvironment  string                     `json:"current_environment"`
	Environments        map[string]evalEnvContext  `json:"environments"`
	KafkaClusterContext evalKafkaClusterContext    `json:"kafka_cluster_context"`
	GlobalAPIKeys       map[string]json.RawMessage `json:"global_api_keys"`
}

type evalEnvContext struct {
	CurrentFlinkComputePool string `json:"current_flink_compute_pool"`
}

type evalKafkaClusterContext struct {
	KafkaEnvironmentContexts map[string]struct {
		ActiveKafka string `json:"active_kafka"`
	} `json:"kafka_environment_contexts"`
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

// GradeConfigIntegrity fails if the config is missing, unparseable, or structurally incoherent. An
// empty current_context is a valid logged-out state (not corruption); only a non-empty
// current_context that names a missing context indicates a torn/incomplete config.
func GradeConfigIntegrity(homeDir string) error {
	c, err := loadConfig(homeDir)
	if err != nil {
		return err
	}
	if c.CurrentContext != "" {
		if _, ok := c.Contexts[c.CurrentContext]; !ok {
			return fmt.Errorf("current_context %q missing from contexts", c.CurrentContext)
		}
	}
	return nil
}

// soleContext returns the name of the single context the eval always produces (one login identity).
func soleContext(homeDir string) (string, error) {
	c, err := loadConfig(homeDir)
	if err != nil {
		return "", err
	}
	if len(c.Contexts) != 1 {
		return "", fmt.Errorf("expected exactly one context, found %d", len(c.Contexts))
	}
	for name := range c.Contexts {
		return name, nil
	}
	return "", fmt.Errorf("no context found")
}

// CredsCleared reports whether the sole context's auth token has been cleared (logged out). Logout
// keeps the context_states entry and only blanks its token, so a missing entry is an error.
func CredsCleared(homeDir string) (bool, error) {
	c, err := loadConfig(homeDir)
	if err != nil {
		return false, err
	}
	name, err := soleContext(homeDir)
	if err != nil {
		return false, err
	}
	state, ok := c.ContextStates[name]
	if !ok {
		return false, fmt.Errorf("context_states missing entry for context %q", name)
	}
	return state.AuthToken == "", nil
}

// ReadActiveKafkaCluster returns the active kafka cluster id for the sole context under the given env.
func ReadActiveKafkaCluster(homeDir, env string) (string, error) {
	c, err := loadConfig(homeDir)
	if err != nil {
		return "", err
	}
	name, err := soleContext(homeDir)
	if err != nil {
		return "", err
	}
	return c.Contexts[name].KafkaClusterContext.KafkaEnvironmentContexts[env].ActiveKafka, nil
}

// GlobalAPIKeyPresent reports whether the sole context's global_api_keys map contains key.
func GlobalAPIKeyPresent(homeDir, key string) (bool, error) {
	c, err := loadConfig(homeDir)
	if err != nil {
		return false, err
	}
	name, err := soleContext(homeDir)
	if err != nil {
		return false, err
	}
	_, ok := c.Contexts[name].GlobalAPIKeys[key]
	return ok, nil
}

// ReadCurrentFlinkComputePool returns the current flink compute pool for the sole context under env.
func ReadCurrentFlinkComputePool(homeDir, env string) (string, error) {
	c, err := loadConfig(homeDir)
	if err != nil {
		return "", err
	}
	name, err := soleContext(homeDir)
	if err != nil {
		return "", err
	}
	return c.Contexts[name].Environments[env].CurrentFlinkComputePool, nil
}
