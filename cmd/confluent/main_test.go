package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/confluentinc/cli/mock"
)

func TestAddCommands_ShownInHelpUsage_CCloud(t *testing.T) {
	req := require.New(t)

	cfg := v2.AuthenticatedConfigMock()
	cfg.CLIName = "ccloud"
	ver := pversion.NewVersion("ccloud", "Confluent Cloud CLI", "https://confluent.cloud; support@confluent.io", "1.2.3", "abc1234", "01/23/45", "CI")
	root, err := cmd.NewConfluentCommand("ccloud", cfg, cfg.Logger, ver, mock.NewDummyAnalyticsMock())
	req.NoError(err)

	output, err := pcmd.ExecuteCommand(root.Command, "help")
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

	cfg := v2.AuthenticatedConfigMock()
	ver := pversion.NewVersion("ccloud", "Confluent Cloud CLI", "https://confluent.cloud; support@confluent.io", "1.2.3", "abc1234", "01/23/45", "CI")
	root, err := cmd.NewConfluentCommand("confluent", cfg, cfg.Logger, ver, mock.NewDummyAnalyticsMock())
	req.NoError(err)

	output, err := pcmd.ExecuteCommand(root.Command, "help")
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
