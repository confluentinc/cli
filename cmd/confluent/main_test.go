package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

func TestAddCommands_ShownInHelpUsage_CCloud(t *testing.T) {
	req := require.New(t)

	cfg := config.AuthenticatedConfigMock()
	cfg.CLIName = "ccloud"
	root, err := cmd.NewConfluentCommand("ccloud", cfg, cfg.Logger)
	req.NoError(err)

	output, err := pcmd.ExecuteCommand(root, "help")
	req.NoError(err)
	req.Contains(output, "kafka")
	//Hidden: req.Contains(output, "ksql")
	req.Contains(output, "environment")
	req.Contains(output, "service-account")
	req.Contains(output, "api-key")
	req.Contains(output, "login")
	req.Contains(output, "logout")
	req.Contains(output, "help")
	req.Contains(output, "version")
	req.Contains(output, "completion")
}

func TestAddCommands_ShownInHelpUsage_Confluent(t *testing.T) {
	req := require.New(t)

	cfg := config.AuthenticatedConfigMock()
	root, err := cmd.NewConfluentCommand("confluent", cfg, cfg.Logger)
	req.NoError(err)

	output, err := pcmd.ExecuteCommand(root, "help")
	req.NoError(err)
	req.NotContains(output, "kafka")
	req.NotContains(output, "ksql")
	req.NotContains(output, "Manage and select")
	req.NotContains(output, "service-account")
	req.NotContains(output, "api-key")
	req.Contains(output, "login")
	req.Contains(output, "logout")
	req.Contains(output, "help")
	req.Contains(output, "version")
	req.Contains(output, "completion")
}
