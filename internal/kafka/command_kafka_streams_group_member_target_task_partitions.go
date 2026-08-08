package kafka

import (
	"github.com/spf13/cobra"
)

func (c *streamsGroupCommand) newStreamsGroupMemberTargetAssignmentTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subtopology",
		Short: "Manage Kafka streams group member target subtopologies.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentTaskPartitionsDescribeCommand())

	return cmd
}
