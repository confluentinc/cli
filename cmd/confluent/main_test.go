package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/terminal"
	cliVersion "github.com/confluentinc/cli/internal/pkg/version"
)

func TestAddCommands_ShownInHelpUsage(t *testing.T) {
	req := require.New(t)

	logger := log.New()
	cfg := config.New(&config.Config{
		Logger: logger,
	})

	version := cliVersion.NewVersion("1.2.3", "abc1234", "01/23/45", "CI")

	// Cloud test
	root := cmd.NewConfluentCommand(cfg, version, logger, "cloud")
	prompt := terminal.NewPrompt(os.Stdin)
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		prompt.SetOutput(cmd.OutOrStderr())
	}

	output, err := terminal.ExecuteCommand(root, "help")
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

	// RBAC test
	root = cmd.NewConfluentCommand(cfg, version, logger, "rbac")
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		prompt.SetOutput(cmd.OutOrStderr())
	}

	output, err = terminal.ExecuteCommand(root, "help")
	req.NoError(err)
	req.NotContains(output, "kafka")
	req.NotContains(output, "ksql")
	req.NotContains(output, "environment")
	req.NotContains(output, "service-account")
	req.NotContains(output, "api-key")
	req.Contains(output, "login")
	req.Contains(output, "logout")
	req.Contains(output, "help")
	req.Contains(output, "version")
	req.Contains(output, "completion")
}
