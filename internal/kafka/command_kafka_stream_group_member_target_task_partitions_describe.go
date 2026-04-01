package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamGroupCommand) newStreamGroupMemberTargetAssignmentTaskPartitionsDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <member>",
		Short:             "Describe stream group member target assignment task partitions",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStreamGroupArgs),
		RunE:              c.streamGroupMemberTargetAssignmentTaskPartitionsDescribe,
	}

	cmd.Flags().String("group", "", "Group Id.")
	cmd.Flags().String("subtopology", "", "Subtopology Id.")
	cmd.Flags().String("assignment-type", "", "Assignments type (active, standby, warmup).")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))
	cobra.CheckErr(cmd.MarkFlagRequired("subtopology"))
	cobra.CheckErr(cmd.MarkFlagRequired("assignment-type"))

	return cmd
}

func (c *streamGroupCommand) streamGroupMemberTargetAssignmentTaskPartitionsDescribe(cmd *cobra.Command, args []string) error {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	memberId := args[0]

	subtopologyId, err := cmd.Flags().GetString("subtopology")
	if err != nil {
		return err
	}

	assignmentsType, err := cmd.Flags().GetString("assignment-type")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	taskPartitions, err := kafkaREST.CloudClientInternal.
		GetKafkaStreamsGroupMemberTargetAssignmentTaskPartitions(groupId, memberId, assignmentsType, subtopologyId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamTaskOut{
		Kind:          taskPartitions.GetKind(),
		SubtopologyId: taskPartitions.GetSubtopologyId(),
		PartitionIds:  taskPartitions.GetPartitionIds(),
	})

	return table.Print()
}
