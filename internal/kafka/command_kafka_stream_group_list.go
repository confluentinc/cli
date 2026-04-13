package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamGroupCommand) newStreamGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List kafka stream groups.",
		Args:        cobra.NoArgs,
		RunE:        c.listStreamGroup,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *streamGroupCommand) listStreamGroup(cmd *cobra.Command, _ []string) error {
	groups, err := c.getStreamGroups(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, streamGroup := range groups {
		list.Add(&streamGroupOut{
			ClusterId:             streamGroup.GetClusterId(),
			GroupId:               streamGroup.GetGroupId(),
			State:                 streamGroup.GetState(),
			MemberCount:           streamGroup.GetMemberCount(),
			SubtopologyCount:      streamGroup.GetSubtopologyCount(),
			GroupEpoch:            streamGroup.GetGroupEpoch(),
			TopologyEpoch:         streamGroup.GetTopologyEpoch(),
			TargetAssignmentEpoch: streamGroup.GetTargetAssignmentEpoch(),
			Members:               streamGroup.Members.GetRelated(),
			Subtopologies:         streamGroup.Subtopologies.GetRelated(),
		})
	}
	return list.Print()
}

func (c *streamGroupCommand) getStreamGroups(cmd *cobra.Command) ([]kafkarestv3.StreamsGroupData, error) {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	topics, err := kafkaREST.CloudClient.ListKafkaStreamsGroup()
	if err != nil {
		return nil, err
	}

	return topics.Data, nil
}
