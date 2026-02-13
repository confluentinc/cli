package kafka

import (
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupMemberTargetAssignmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-group-member-target-assignment",
		Short: "Manage Kafka stream group member target assignments.",
	}

	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentDescribeCommand())
	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentTaskListCommand())

	return cmd
}
