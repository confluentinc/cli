package kafka

import (
	"github.com/spf13/cobra"
)

type streamsGroupMemberOut struct {
	Kind          string `human:"Kind" serialized:"kind"`
	ClusterId     string `human:"Cluster Id" serialized:"cluster_id"`
	GroupId       string `human:"Group Id" serialized:"group_id"`
	MemberId      string `human:"Member Id" serialized:"member_id"`
	ProcessId     string `human:"Process Id" serialized:"process_id"`
	ClientId      string `human:"Client Id" serialized:"client_id"`
	InstanceId    string `human:"Instance Id" serialized:"instance_id"`
	MemberEpoch   int32  `human:"Member Epoch" serialized:"member_epoch"`
	TopologyEpoch int32  `human:"Topology Epoch" serialized:"topology_epoch"`
	IsClassic     bool   `human:"Is Classic" serialized:"is_classic"`
	Assignments   string `human:"Assignments" serialized:"assignments"`
	TargetAssign  string `human:"Target Assignment" serialized:"target_assignment"`
}

func (c *streamsGroupCommand) newStreamsGroupMemberCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage Kafka stream groups members.",
	}

	cmd.AddCommand(c.newStreamsGroupMemberDescribeCommand())
	cmd.AddCommand(c.newStreamsGroupMemberListCommand())

	return cmd
}
