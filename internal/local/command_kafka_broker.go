package local

import (
	"github.com/spf13/cobra"
)

func (c *command) newKafkaBrokerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broker",
		Short: "Manage local Kafka brokers.",
	}

	cmd.AddCommand(c.newKafkaBrokerConfigurationCommand())
	cmd.AddCommand(c.newKafkaBrokerDescribeCommand())
	cmd.AddCommand(c.newKafkaBrokerListCommand())

	return cmd
}
