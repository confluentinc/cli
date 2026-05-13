package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupMemberListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka streams group members.",
		Args:  cobra.NoArgs,
		RunE:  c.listStreamsGroupMembers,
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

func (c *streamsGroupCommand) listStreamsGroupMembers(cmd *cobra.Command, _ []string) error {
	members, err := c.getStreamsGroupMembers(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, member := range members {
		list.Add(&streamsGroupMemberOut{
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
	}
	return list.Print()
}

func (c *streamsGroupCommand) getStreamsGroupMembers(cmd *cobra.Command) ([]kafkarestv3.StreamsGroupMemberData, error) {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := kafkaREST.CloudClient.ListKafkaStreamsGroupMembers(groupId)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
