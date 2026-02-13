package kafka

import (
	"github.com/spf13/cobra"
)

type streamTaskOut struct {
	Kind          string  `human:"Kind" serialized:"kind"`
	SubtopologyId string  `human:"Subtopology Id" serialized:"subtopology_id"`
	PartitionIds  []int32 `human:"Partition Ids" serialized:"partition_ids"`
}

func (c *consumerCommand) newStreamGroupMemberTaskPartitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-group-member-task-partitions",
		Short: "Manage Kafka stream group member task partitions.",
	}

	cmd.AddCommand(c.newStreamGroupMemberTaskPartitionsDescribeCommand())

	return cmd
}
