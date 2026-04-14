package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe a Kafka streams group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStreamsGroupArgs),
		RunE:              c.streamsGroupDescribe,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *streamsGroupCommand) streamsGroupDescribe(cmd *cobra.Command, args []string) error {
	groupId := args[0]

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	streamsGroup, err := kafkaREST.CloudClient.GetKafkaStreamsGroup(groupId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamsGroupOut{
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

	return table.Print()
}
