package ksql

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/shell/completer"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
	prerunner       pcmd.PreRunner
	serverCompleter completer.ServerSideCompleter
	analyticsClient analytics.Client
}

// New returns the default command object for interacting with KSQL.
func New(cliName string, prerunner pcmd.PreRunner, serverCompleter completer.ServerSideCompleter, analyticsClient analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewCLICommand(
		&cobra.Command{
			Use:   "ksql",
			Short: "Manage ksqlDB applications.",
		}, prerunner)
	cmd := &command{
		CLICommand:      cliCmd,
		prerunner:       prerunner,
		serverCompleter: serverCompleter,
		analyticsClient: analyticsClient,
	}
	cmd.init(cliName)
	return cmd.Command
}

func (c *command) init(cliName string) {
	if cliName == "ccloud" {
		clusterCmd := NewClusterCommand(c.prerunner, c.analyticsClient)
		c.AddCommand(clusterCmd.Command)
		c.serverCompleter.AddCommand(clusterCmd)
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
