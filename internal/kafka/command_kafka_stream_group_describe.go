package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamGroupCommand) newStreamGroupDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe stream group",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStreamGroupArgs),
		RunE:              c.streamGroupDescribe,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *streamGroupCommand) streamGroupDescribe(cmd *cobra.Command, args []string) error {
	groupId := args[0]

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	streamGroup, err := kafkaREST.CloudClient.GetKafkaStreamGroup(groupId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamGroupOut{
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

	return table.Print()
}
