package local

import (
	"github.com/spf13/cobra"
)

func (c *command) newKafkaClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kafka cluster in Confluent Local.",
	}

	cmd.AddCommand(c.newKafkaClusterConfigurationCommand())

	return cmd
}
