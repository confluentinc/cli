package kafka

import (
	"github.com/spf13/cobra"
)

func (c *brokerCommand) newConfigurationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage Kafka broker configurations.",
	}

	cmd.AddCommand(c.newConfigurationListCommand())
	cmd.AddCommand(c.newConfigurationUpdateCommand())

	return cmd
}
