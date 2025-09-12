package kafka

import (
	"github.com/spf13/cobra"
)

func (c *shareCommand) newGroupConsumerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage Kafka share group consumers.",
	}

	// Only cloud support for now
	cmd.AddCommand(c.newGroupConsumerListCommand())

	return cmd
}
