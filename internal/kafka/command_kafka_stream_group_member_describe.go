package kafka

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupMemberDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <member>",
		Short:             "Describe stream group member",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStreamGroupArgs),
		RunE:              c.streamGroupMemberDescribe,
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

func (c *consumerCommand) streamGroupMemberDescribe(cmd *cobra.Command, args []string) error {
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

	member, err := kafkaREST.CloudClientInternal.GetKafkaStreamGroupMember(groupId, memberId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamGroupMemberOut{
		Kind:          member.GetKind(),
		ClusterId:     member.GetClusterId(),
		GroupId:       member.GetGroupId(),
		MemberId:      member.GetMemberId(),
		ProcessId:     member.GetProcessId(),
		ClientId:      member.GetClientId(),
		InstanceId:    member.GetInstanceId(),
		MemberEpoch:   member.GetMemberEpoch(),
		TopologyEpoch: member.GetTopologyEpoch(),
		IsClassic:     member.GetIsClassic(),
		Assignments:   member.Assignments.GetRelated(),
		TargetAssign:  member.TargetAssignment.GetRelated(),
	})

	return table.Print()
}
