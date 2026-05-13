package kafka

import (
	"github.com/spf13/cobra"
)

func (c *streamsGroupCommand) newStreamsGroupMemberTargetAssignmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target-assignment",
		Short: "Manage Kafka streams group member target assignments.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentDescribeCommand())
	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentTaskListCommand())
	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentTaskPartitionsCommand())

	return cmd
}
