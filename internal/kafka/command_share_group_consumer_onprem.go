package kafka

import (
	"github.com/spf13/cobra"
)

func (c *shareGroupCommand) newConsumerCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage Kafka share group consumers.",
	}

	cmd.AddCommand(c.newConsumerListCommandOnPrem())

	return cmd
}
