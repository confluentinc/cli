package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.CLICommand
	prerunner   pcmd.PreRunner
}

// New returns the default command object for interacting with KSQL.
func New(prerunner pcmd.PreRunner, config *config.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "ksql",
			Short: "Manage KSQL.",
		},
		config, prerunner)
	cmd := &command{
		CLICommand: cliCmd,
		prerunner:   prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.Config, c.prerunner))
}
