package kafka

import (
	"github.com/spf13/cobra"
)

func (c *streamsGroupCommand) newStreamsGroupMemberTargetAssignmentTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member-target-assignment-task-partitions",
		Short: "Manage stream group target task partitions.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentTaskPartitionsDescribeCommand())

	return cmd
}
