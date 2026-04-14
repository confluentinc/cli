package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka stream groups.",
		Args:  cobra.NoArgs,
		RunE:  c.listStreamsGroup,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *streamsGroupCommand) listStreamsGroup(cmd *cobra.Command, _ []string) error {
	groups, err := c.getStreamsGroups(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, streamsGroup := range groups {
		list.Add(&streamsGroupOut{
			ClusterId:             streamsGroup.GetClusterId(),
			GroupId:               streamsGroup.GetGroupId(),
			State:                 streamsGroup.GetState(),
			MemberCount:           streamsGroup.GetMemberCount(),
			SubtopologyCount:      streamsGroup.GetSubtopologyCount(),
			GroupEpoch:            streamsGroup.GetGroupEpoch(),
			TopologyEpoch:         streamsGroup.GetTopologyEpoch(),
			TargetAssignmentEpoch: streamsGroup.GetTargetAssignmentEpoch(),
			Members:               streamsGroup.Members.GetRelated(),
			Subtopologies:         streamsGroup.Subtopologies.GetRelated(),
		})
	}
	return list.Print()
}

func (c *streamsGroupCommand) getStreamsGroups(cmd *cobra.Command) ([]kafkarestv3.StreamsGroupData, error) {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	groups, err := kafkaREST.CloudClient.ListKafkaStreamsGroup()
	if err != nil {
		return nil, err
	}

	return groups.Data, nil
}
