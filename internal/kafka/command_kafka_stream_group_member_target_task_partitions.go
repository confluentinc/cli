package kafka

import (
	"github.com/spf13/cobra"
)

func (c *streamsGroupCommand) newStreamsGroupMemberTargetAssignmentTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "member-target-assignment-task-partitions",
		Aliases: []string{"mtatp"},
		Short:   "Manage Kafka streams group member target task partitions.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentTaskPartitionsDescribeCommand())

	return cmd
}
