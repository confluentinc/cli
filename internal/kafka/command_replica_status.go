package kafka

import (
	"github.com/spf13/cobra"
)

func (c *replicaCommand) newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Manage Kafka replica statuses.",
	}

	cmd.AddCommand(c.newStatusListCommand())

	return cmd
}
