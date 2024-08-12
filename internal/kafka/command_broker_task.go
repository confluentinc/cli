package kafka

import (
	"github.com/spf13/cobra"
)

func (c *brokerCommand) newTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage broker tasks.",
	}

	cmd.AddCommand(c.newTaskListCommand())

	return cmd
}
