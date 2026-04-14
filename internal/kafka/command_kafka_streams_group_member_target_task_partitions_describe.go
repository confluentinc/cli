package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupMemberTargetAssignmentTaskPartitionsDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <member>",
		Short: "Describe Kafka streams group member target task partitions.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.streamsGroupMemberTargetAssignmentTaskPartitionsDescribe,
	}

	cmd.Flags().String("group", "", "Group Id.")
	cmd.Flags().String("subtopology", "", "Subtopology Id.")
	cmd.Flags().String("assignment-type", "", "Assignment type (active, standby, warmup).")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))
	cobra.CheckErr(cmd.MarkFlagRequired("subtopology"))
	cobra.CheckErr(cmd.MarkFlagRequired("assignment-type"))

	pcmd.RegisterFlagCompletionFunc(cmd, "assignment-type", func(_ *cobra.Command, _ []string) []string {
		return []string{"active", "standby", "warmup"}
	})

	return cmd
}

func (c *streamsGroupCommand) streamsGroupMemberTargetAssignmentTaskPartitionsDescribe(cmd *cobra.Command, args []string) error {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	memberId := args[0]

	subtopologyId, err := cmd.Flags().GetString("subtopology")
	if err != nil {
		return err
	}

	assignmentType, err := cmd.Flags().GetString("assignment-type")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	taskPartitions, err := kafkaREST.CloudClient.
		GetKafkaStreamsGroupMemberTargetAssignmentTaskPartitions(groupId, memberId, assignmentType, subtopologyId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamsTaskOut{
		Kind:          taskPartitions.GetKind(),
		SubtopologyId: taskPartitions.GetSubtopologyId(),
		PartitionIds:  taskPartitions.GetPartitionIds(),
	})

	return table.Print()
}
