package kafka

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/spf13/cobra"
)

type streamGroupOut struct {
	Kind                  string `human:"Kind" serialized:"kind"`
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

func (c *consumerCommand) newStreamGroupCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "stream-group",
		Short:       "Manage Kafka stream groups.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	cmd.AddCommand(c.newStreamGroupDescribeCommand())
	cmd.AddCommand(c.newStreamGroupListCommand())
	cmd.AddCommand(c.newStreamGroupMemberCommand())
	cmd.AddCommand(c.newStreamGroupMemberAssignmentCommand())
	cmd.AddCommand(c.newStreamGroupMemberTaskPartitionsCommand())
	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentCommand())
	cmd.AddCommand(c.newStreamGroupMemberTargetAssignmentTaskPartitionsCommand())
	cmd.AddCommand(c.newStreamGroupSubtopologyCommand())

	return cmd
}

func (c *consumerCommand) validStreamGroupArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteStreamGroups(cmd, c.AuthenticatedCLICommand)
}
