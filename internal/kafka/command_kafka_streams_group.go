package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type streamsGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type streamsGroupOut struct {
	ClusterId             string `human:"Cluster Id" serialized:"cluster_id"`
	GroupId               string `human:"Group Id" serialized:"group_id"`
	State                 string `human:"State" serialized:"state"`
	MemberCount           int32  `human:"Member Count" serialized:"member_count"`
	SubtopologyCount      int32  `human:"Subtopology Count" serialized:"subtopology_count"`
	GroupEpoch            int32  `human:"Group Epoch" serialized:"group_epoch"`
	TopologyEpoch         int32  `human:"Topology Epoch" serialized:"topology_epoch"`
	TargetAssignmentEpoch int32  `human:"Target Assignment Epoch" serialized:"target_assignment_epoch"`
	Members               string `human:"Members" serialized:"members"`
	Subtopologies         string `human:"Subtopologies" serialized:"subtopologies"`
}

func newStreamsGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "streams-group",
		Short:       "Manage Kafka streams groups.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &streamsGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newStreamsGroupDescribeCommand())
	cmd.AddCommand(c.newStreamsGroupListCommand())
	cmd.AddCommand(c.newStreamsGroupMemberAssignmentCommand())
	cmd.AddCommand(c.newStreamsGroupMemberCommand())
	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentCommand())
	cmd.AddCommand(c.newStreamsGroupMemberTargetAssignmentTaskPartitionsCommand())
	cmd.AddCommand(c.newStreamsGroupMemberTaskPartitionsCommand())
	cmd.AddCommand(c.newStreamsGroupSubtopologyCommand())

	return cmd
}

func (c *streamsGroupCommand) validStreamsGroupArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteStreamsGroups(cmd, c.AuthenticatedCLICommand)
}
