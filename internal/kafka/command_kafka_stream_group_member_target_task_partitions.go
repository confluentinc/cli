package kafka

import (
	"github.com/spf13/cobra"
)

func (c *streamGroupCommand) newStreamGroupMemberTargetAssignmentTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member-target-assignment-task-partitions",
		Short: "Manage stream group target task partitions.",
	}

	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentTaskPartitionsDescribeCommand())

	return cmd
}
