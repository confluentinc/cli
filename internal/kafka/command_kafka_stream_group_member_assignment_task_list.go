package kafka

import (
	kafkarestv3Internal "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafkarest/v3"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupMemberAssignmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List kafka stream group member assignment tasks.",
		Args:        cobra.NoArgs,
		RunE:        c.listStreamGroupMemberAssignmentTasks,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().String("group", "", "Group Id.")
	cmd.Flags().String("member", "", "Member Id.")
	cmd.Flags().String("assignment", "", "Assignment type (active, standby, warmup).")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))
	cobra.CheckErr(cmd.MarkFlagRequired("member"))
	cobra.CheckErr(cmd.MarkFlagRequired("assignment"))

	return cmd
}

func (c *consumerCommand) listStreamGroupMemberAssignmentTasks(cmd *cobra.Command, _ []string) error {
	tasks, err := c.getStreamGroupMemberAssignmentTasks(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, task := range tasks {
		list.Add(&streamTaskOut{
			Kind:          task.GetKind(),
			SubtopologyId: task.GetSubtopologyId(),
			PartitionIds:  task.GetPartitionIds(),
		})
	}

	return list.Print()
}

func (c *consumerCommand) getStreamGroupMemberAssignmentTasks(cmd *cobra.Command) ([]kafkarestv3Internal.StreamsTaskData, error) {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}

	memberId, err := cmd.Flags().GetString("member")
	if err != nil {
		return nil, err
	}

	assignmentType, err := cmd.Flags().GetString("assignment")
	if err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := kafkaREST.CloudClientInternal.ListKafkaStreamsGroupMemberAssignmentTasks(groupId, memberId, assignmentType)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
