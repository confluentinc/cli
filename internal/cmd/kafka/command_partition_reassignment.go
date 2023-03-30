package kafka

import "github.com/spf13/cobra"

func (c *partitionCommand) newReassignmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reassignment",
		Short: "Manage ongoing partition reassignments.",
	}

	cmd.AddCommand(c.newReassignmentListCommand())

	return cmd
}
