package local

import (
	"github.com/spf13/cobra"
)

func (c *command) newKafkaBrokerConfigurationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage Kafka broker configurations in Confluent Local.",
	}

	cmd.AddCommand(c.newKafkaBrokerConfigurationListCommand())
	cmd.AddCommand(c.newKafkaBrokerConfigurationUpdateCommand())

	return cmd
}
