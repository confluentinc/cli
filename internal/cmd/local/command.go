package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/local"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

type LocalCommand struct {
	*cmd.CLICommand
	ch local.ConfluentHome
	cc local.ConfluentCurrent
}

func NewLocalCommand(command *cobra.Command, prerunner cmd.PreRunner) *LocalCommand {
	return &LocalCommand{
		CLICommand: cmd.NewAnonymousCLICommand(command, prerunner),
		ch:         local.NewConfluentHomeManager(),
		cc:         local.NewConfluentCurrentManager(),
	}
}

func NewCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "local-v2 [command]",
			Short: "Manage a local Confluent Platform development environment.",
		}, prerunner)

	c.AddCommand(NewCurrentCommand(prerunner))
	c.AddCommand(NewDemoCommand(prerunner))
	c.AddCommand(NewDestroyCommand(prerunner))
	c.AddCommand(NewServicesCommand(prerunner))
	c.AddCommand(NewVersionCommand(prerunner))

	return c.Command
}
