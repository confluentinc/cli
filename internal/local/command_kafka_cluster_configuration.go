package local

import (
	"github.com/spf13/cobra"
)

func (c *Command) newKafkaClusterConfigurationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage Kafka cluster configurations in Confluent Local.",
	}

	cmd.AddCommand(c.newKafkaClusterConfigurationListCommand())

	return cmd
}
