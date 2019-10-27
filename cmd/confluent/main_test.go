package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

func TestAddCommands_ShownInHelpUsage_CCloud(t *testing.T) {
	req := require.New(t)

	logger := log.New()
	cfg := config.New(&config.Config{
		CLIName: "ccloud",
		Logger:  logger,
	})
	req.NoError(cfg.Load())

	version := cliVersion.NewVersion("ccloud", "Confluent Cloud CLI", "https://confluent.cloud; support@confluent.io", "1.2.3", "abc1234", "01/23/45", "CI")

	root, err := cmd.NewConfluentCommand("ccloud", cfg, version, logger)
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

	logger := log.New()
	cfg := config.New(&config.Config{
		CLIName: "confluent",
		Logger:  logger,
	})
	req.NoError(cfg.Load())

	version := cliVersion.NewVersion("confluent", "Confluent CLI", "https://confluent.io; support@confluent.io", "1.2.3", "abc1234", "01/23/45", "CI")

	root, err := cmd.NewConfluentCommand("confluent", cfg, version, logger)
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
	req.Contains(output, "iam")
}
