package kafka

import (
	"github.com/spf13/cobra"
)

type streamsTaskOut struct {
	Kind          string  `human:"Kind" serialized:"kind"`
	SubtopologyId string  `human:"Subtopology Id" serialized:"subtopology_id"`
	PartitionIds  []int32 `human:"Partition Ids" serialized:"partition_ids"`
}

func (c *streamsGroupCommand) newStreamsGroupMemberTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "member-task-partitions",
		Aliases: []string{"mtp"},
		Short:   "Manage Kafka streams group member task partitions.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberTaskPartitionsDescribeCommand())

	return cmd
}
