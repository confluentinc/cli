package kafka

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupMemberTargetAssignmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <member>",
		Short:             "Describe stream group member target assignment",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStreamGroupArgs),
		RunE:              c.streamGroupMemberTargetAssignmentDescribe,
	}

	cmd.Flags().String("group", "", "Group Id.")
	cmd.Flags().String("member", "", "Member Id.")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))
	cobra.CheckErr(cmd.MarkFlagRequired("member"))

	return cmd
}

func (c *consumerCommand) streamGroupMemberTargetAssignmentDescribe(cmd *cobra.Command, args []string) error {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	memberId, err := cmd.Flags().GetString("member")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	assignment, err := kafkaREST.CloudClientInternal.GetKafkaStreamGroupMemberTargetAssignment(groupId, memberId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamGroupMemberAssignmentOut{
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
