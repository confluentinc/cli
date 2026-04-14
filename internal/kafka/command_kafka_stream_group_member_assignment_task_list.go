package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupMemberAssignmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka streams group member assignment tasks.",
		Args:  cobra.NoArgs,
		RunE:  c.listStreamsGroupMemberAssignmentTasks,
	}

	cmd.Flags().String("group", "", "Group Id.")
	cmd.Flags().String("member", "", "Member Id.")
	cmd.Flags().String("assignment-type", "", "Assignment type (active, standby, warmup).")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))
	cobra.CheckErr(cmd.MarkFlagRequired("member"))
	cobra.CheckErr(cmd.MarkFlagRequired("assignment-type"))

	pcmd.RegisterFlagCompletionFunc(cmd, "assignment-type", func(_ *cobra.Command, _ []string) []string {
		return []string{"active", "standby", "warmup"}
	})

	return cmd
}

func (c *streamsGroupCommand) listStreamsGroupMemberAssignmentTasks(cmd *cobra.Command, _ []string) error {
	tasks, err := c.getStreamsGroupMemberAssignmentTasks(cmd)
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

func (c *streamsGroupCommand) getStreamsGroupMemberAssignmentTasks(cmd *cobra.Command) ([]kafkarestv3.StreamsTaskData, error) {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}

	memberId, err := cmd.Flags().GetString("member")
	if err != nil {
		return nil, err
	}

	assignmentType, err := cmd.Flags().GetString("assignment-type")
	if err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := kafkaREST.CloudClient.ListKafkaStreamsGroupMemberAssignmentTasks(groupId, memberId, assignmentType)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
