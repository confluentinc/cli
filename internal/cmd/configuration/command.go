package configuration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.CLICommand
	config *config.Config
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config"},
		Short:   "Manage CLI configuration fields.",
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		config:     cfg,
	}

	cmd.AddCommand(c.newSetCommand())
	//cmd.AddCommand(c.newListCommand())

	return cmd
}
