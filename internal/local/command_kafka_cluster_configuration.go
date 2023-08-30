package local

import (
	"github.com/spf13/cobra"
)

func (c *command) newKafkaClusterConfigurationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage Kafka cluster configurations in Confluent Local.",
	}

	cmd.AddCommand(c.newKafkaClusterConfigurationListCommand())
	cmd.AddCommand(c.newKafkaClusterConfigurationUpdateCommand())

	return cmd
}
