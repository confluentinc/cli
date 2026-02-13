package kafka

import (
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupMemberTargetAssignmentTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-group-member-target-assignment-task-partitions",
		Short: "Manage Kafka stream group member target assignment task partitions.",
	}

	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentTaskPartitionsDescribeCommand())

	return cmd
}
