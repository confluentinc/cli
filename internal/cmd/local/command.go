package local

import (
	"runtime"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/local"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

type Command struct {
	*cmd.CLICommand
	ch local.ConfluentHome
	cc local.ConfluentCurrent
}

func NewLocalCommand(command *cobra.Command, prerunner cmd.PreRunner) *Command {
	return &Command{
		CLICommand: cmd.NewAnonymousCLICommand(command, prerunner),
		ch:         local.NewConfluentHomeManager(),
		cc:         local.NewConfluentCurrentManager(),
	}
}

func New(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "local",
			Short: "Manage a local Confluent Platform development environment.",
			Long:  "Use the \"confluent local\" commands to try out Confluent Platform by running a single-node instance locally on your machine. Keep in mind, these commands require Java to run.",
			Args:  cobra.NoArgs,
		}, prerunner)

	if runtime.GOOS == "windows" {
		c.Hidden = true
	}

	c.AddCommand(NewCurrentCommand(prerunner))
	c.AddCommand(NewDestroyCommand(prerunner))
	c.AddCommand(NewServicesCommand(prerunner))
	c.AddCommand(NewVersionCommand(prerunner))

	return c.Command
}
