package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with KSQL.
func New(cliName string, prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "ksql",
			Short: "Manage ksqlDB applications.",
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init(cliName)
	return cmd.Command
}

func (c *command) init(cliName string) {
	if cliName == "ccloud" {
		c.AddCommand(NewClusterCommand(c.prerunner))
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
