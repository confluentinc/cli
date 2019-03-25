package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
	cliVersion "github.com/confluentinc/cli/version"
)

func TestAddCommands_ShownInHelpUsage(t *testing.T) {
	req := require.New(t)

	logger := log.New()
	cfg := shared.NewConfig(&shared.Config{
		Logger: logger,
	})

	version := cliVersion.NewVersion("1.2.3", "abc1234", "01/23/45", "CI")
	root := BuildCommand(cfg, version, logger)
	prompt := command.NewTerminalPrompt(os.Stdin)
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		prompt.SetOutput(cmd.OutOrStderr())
	}

	output, err := command.ExecuteCommand(root, "help")
	req.NoError(err)
	req.Contains(output, "kafka")
	//Hidden: req.Contains(output, "ksql")
	req.Contains(output, "environment")
}
