package kafka

import (
	"github.com/spf13/cobra"
)

func (c *shareGroupCommand) newConsumerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage Kafka share group consumers.",
	}

	cmd.AddCommand(c.newConsumerListCommand())

	return cmd
}
