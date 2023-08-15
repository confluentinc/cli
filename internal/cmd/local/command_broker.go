package local

import (
	"github.com/spf13/cobra"
)

func (c *Command) newBrokerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broker",
		Short: "Manage brokers in Confluent Local cluster",
	}

	cmd.AddCommand(c.newBrokerListCommand())
	// cmd.AddCommand(c.newBrokerDescribeCommand())
	// cmd.AddCommand(c.newBrokerUpdateCommand())

	return cmd
}
