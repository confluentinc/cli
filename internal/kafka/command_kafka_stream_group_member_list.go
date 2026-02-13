package kafka

import (
	kafkarestv3Internal "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafkarest/v3"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupMemberListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List kafka stream group members.",
		Args:        cobra.NoArgs,
		RunE:        c.listStreamGroupMembers,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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

func (c *consumerCommand) listStreamGroupMembers(cmd *cobra.Command, _ []string) error {
	members, err := c.getStreamGroupMembers(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, member := range members {
		list.Add(&streamGroupMemberOut{
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

func (c *consumerCommand) getStreamGroupMembers(cmd *cobra.Command) ([]kafkarestv3Internal.StreamsGroupMemberData, error) {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := kafkaREST.CloudClientInternal.ListKafkaStreamsGroupMembers(groupId)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
