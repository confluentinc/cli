package local

import (
	"github.com/spf13/cobra"
)

func (c *command) newKafkaTopicConfigurationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage Kafka topic configurations in Confluent Local.",
	}

	cmd.AddCommand(c.newKafkaTopicConfigurationListCommand())

	return cmd
}
