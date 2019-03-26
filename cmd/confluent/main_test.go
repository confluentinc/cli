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
	root := cmd.NewConfluentCommand(cfg, version, logger)
	prompt := terminal.NewPrompt(os.Stdin)
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		prompt.SetOutput(cmd.OutOrStderr())
	}

	output, err := terminal.ExecuteCommand(root, "help")
	req.NoError(err)
	req.Contains(output, "kafka")
	//Hidden: req.Contains(output, "ksql")
	req.Contains(output, "environment")
}
