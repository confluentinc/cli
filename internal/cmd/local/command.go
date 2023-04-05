package local

import (
	"runtime"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/local"
)

type Command struct {
	*pcmd.CLICommand
	ch local.ConfluentHome
	cc local.ConfluentCurrent
}

func NewLocalCommand(cmd *cobra.Command, prerunner pcmd.PreRunner) *Command {
	return &Command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		ch:         local.NewConfluentHomeManager(),
		cc:         local.NewConfluentCurrentManager(),
	}
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "local",
			Short: "Manage a local Confluent Platform development environment.",
			Long:  "Use the `confluent local` commands to try out Confluent Platform by running a single-node instance locally on your machine. Keep in mind, these commands require Java to run.",
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
