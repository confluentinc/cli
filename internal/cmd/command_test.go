package cmd

import (
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"testing"

	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	mockBaseConfig = &config.BaseConfig{
		Params: &config.Params{
			Logger: log.New(),
		},
	}
	mockVersion = new(pversion.Version)
)

func TestHelp_NoConfig(t *testing.T) {
	cfg := &v3.Config{
		BaseConfig: mockBaseConfig,
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"audit-log", "completion", "config", "help", "iam", "kafka", "ksql", "login", "logout", "schema-registry",
		"update", "version",
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func TestHelp_Cloud(t *testing.T) {
	cfg := &v3.Config{
		BaseConfig: mockBaseConfig,
		Contexts: map[string]*v3.Context{
			"context": {
				PlatformName: "confluent.cloud",
			},
		},
		CurrentContext: "context",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"admin", "api-key", "audit-log", "completion", "config", "connector", "connector-catalog", "environment",
		"help", "iam", "init", "kafka", "ksql", "login", "logout", "price", "prompt", "schema-registry",
		"service-account", "shell", "signup", "update", "version",
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func TestHelp_CloudWithAPIKey(t *testing.T) {
	cfg := &v3.Config{
		BaseConfig: mockBaseConfig,
		Contexts: map[string]*v3.Context{
			"context": {
				PlatformName: "confluent.cloud",
				Credential: &v2.Credential{
					CredentialType: v2.APIKey,
				},
			},
		},
		CurrentContext: "context",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"admin", "audit-log", "completion", "config", "help", "init", "kafka", "login", "logout", "update", "version",
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func TestHelp_OnPrem(t *testing.T) {
	cfg := &v3.Config{
		BaseConfig: mockBaseConfig,
		Contexts: map[string]*v3.Context{
			"context": {
				PlatformName: "https://somecompany.com",
			},
		},
		CurrentContext: "context",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"audit-log", "cluster", "completion", "config", "connect", "help", "iam", "kafka", "ksql", "local", "login",
		"logout", "schema-registry", "secret", "update", "version",
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func runWithConfig(cfg *v3.Config) (string, error) {
	cli := NewConfluentCommand(cfg, true, mockVersion)
	return pcmd.ExecuteCommand(cli.Command, "help")
}
