package kafka

import (
	"github.com/spf13/cobra"
)

type streamGroupMemberAssignmentOut struct {
	Kind         string `human:"Kind" serialized:"kind"`
	ClusterId    string `human:"Cluster Id" serialized:"cluster_id"`
	GroupId      string `human:"Group Id" serialized:"group_id"`
	MemberId     string `human:"Member Id" serialized:"member_id"`
	ActiveTasks  string `human:"Active Tasks" serialized:"active_tasks"`
	StandbyTasks string `human:"Standby Tasks" serialized:"standby_tasks"`
	WarmupTasks  string `human:"Warmup Tasks" serialized:"warmup_tasks"`
}

func (c *consumerCommand) newStreamGroupMemberAssignmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-group-member-assignment",
		Short: "Manage Kafka stream group member assignments.",
	}

	cmd.AddCommand(c.newStreamGroupMemberAssignmentDescribeCommand())
	cmd.AddCommand(c.newStreamGroupMemberAssignmentListCommand())

	return cmd
}
