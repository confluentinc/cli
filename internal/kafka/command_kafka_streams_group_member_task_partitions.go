package kafka

import (
	"github.com/spf13/cobra"
)

type streamsTaskOut struct {
	SubtopologyId string  `human:"Subtopology Id" serialized:"subtopology_id"`
	PartitionIds  []int32 `human:"Partition Ids" serialized:"partition_ids"`
}

func (c *streamsGroupCommand) newStreamsGroupMemberTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subtopology",
		Short: "Manage Kafka streams group member assignment subtopologies.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberTaskPartitionsDescribeCommand())

	return cmd
}
