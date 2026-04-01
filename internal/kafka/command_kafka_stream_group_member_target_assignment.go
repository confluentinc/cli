package kafka

import (
	"github.com/spf13/cobra"
)

func (c *streamGroupCommand) newStreamGroupMemberTargetAssignmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member-target-assignment",
		Short: "Manage Kafka stream group member target assignments.",
	}

	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentDescribeCommand())
	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentTaskListCommand())

	return cmd
}
