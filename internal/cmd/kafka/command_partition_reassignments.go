package kafka

import "github.com/spf13/cobra"

func (c *partitionCommand) newReassignmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reassignments",
		Short: "Manage ongoing replica reassignments.",
	}

	cmd.AddCommand(c.newReassignmentsListCommand())

	return cmd
}
