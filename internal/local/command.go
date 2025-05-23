package local

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/local"
)

type command struct {
	*pcmd.CLICommand
	ch local.ConfluentHome
	cc local.ConfluentCurrent
}

func NewLocalCommand(cmd *cobra.Command, prerunner pcmd.PreRunner) *command {
	return &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		ch:         local.NewConfluentHomeManager(),
		cc:         local.NewConfluentCurrentManager(),
	}
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Manage a local Confluent Platform development environment.",
		Long:  "Try out Confluent Platform by running a single-node instance locally on your machine. These commands require Docker to run.",
		Args:  cobra.NoArgs,
	}

	c := NewLocalCommand(cmd, prerunner)

	c.AddCommand(c.newKafkaCommand())

	c.AddCommand(NewCurrentCommand(prerunner))
	c.AddCommand(NewDestroyCommand(prerunner))
	c.AddCommand(NewServicesCommand(cfg, prerunner))
	c.AddCommand(NewVersionCommand(prerunner))

	return cmd
}
