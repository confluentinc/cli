package local

import (
	"github.com/spf13/cobra"
)

func (c *command) newKafkaClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage the local Kafka cluster.",
	}

	cmd.AddCommand(c.newKafkaClusterConfigurationCommand())

	return cmd
}
