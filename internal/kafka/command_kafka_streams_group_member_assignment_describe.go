package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupMemberAssignmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <member>",
		Short: "Describe a Kafka streams group member assignment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.streamsGroupMemberAssignmentDescribe,
	}

	cmd.Flags().String("group", "", "Group Id.")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *streamsGroupCommand) streamsGroupMemberAssignmentDescribe(cmd *cobra.Command, args []string) error {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	memberId := args[0]

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	assignment, err := kafkaREST.CloudClient.GetKafkaStreamsGroupMemberAssignment(groupId, memberId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamsGroupMemberAssignmentOut{
		Kind:         assignment.GetKind(),
		ClusterId:    assignment.GetClusterId(),
		GroupId:      assignment.GetGroupId(),
		MemberId:     assignment.GetMemberId(),
		ActiveTasks:  assignment.ActiveTasks.GetRelated(),
		StandbyTasks: assignment.StandbyTasks.GetRelated(),
		WarmupTasks:  assignment.WarmupTasks.GetRelated(),
	})

	return table.Print()
}
