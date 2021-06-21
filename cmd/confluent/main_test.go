package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func TestAddCommands_ShownInHelpUsage_CCloud(t *testing.T) {
	req := require.New(t)

	ver := pversion.NewVersion("ccloud", "1.2.3", "abc1234", "01/23/45", "CI")
	cli := cmd.NewConfluentCommand("ccloud", true, ver)

	output, err := pcmd.ExecuteCommand(cli.Command, "help")
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

	ver := pversion.NewVersion("ccloud", "1.2.3", "abc1234", "01/23/45", "CI")
	cli := cmd.NewConfluentCommand("confluent", true, ver)

	output, err := pcmd.ExecuteCommand(cli.Command, "help")
	req.NoError(err)
	req.NotContains(output, "Manage and select")
	req.NotContains(output, "service-account")
	req.NotContains(output, "api-key")
	req.Contains(output, "login")
	req.Contains(output, "logout")
	req.Contains(output, "help")
	req.Contains(output, "version")
	req.Contains(output, "completion")
	req.Contains(output, "iam")
	req.Contains(output, "kafka")
	req.Contains(output, "ksql")
	req.Contains(output, "schema-registry")
	req.Contains(output, "connect")
}
