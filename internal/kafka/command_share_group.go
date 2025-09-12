package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type shareGroupOut struct {
	Cluster            string `human:"Cluster" serialized:"cluster"`
	ShareGroup         string `human:"Share Group" serialized:"share_group"`
	Coordinator        string `human:"Coordinator" serialized:"coordinator"`
	State              string `human:"State" serialized:"state"`
	ConsumerCount      int32  `human:"Consumer Count" serialized:"consumer_count"`
	PartitionCount     int32  `human:"Partition Count" serialized:"partition_count"`
	TopicSubscriptions string `human:"Topic Subscriptions" serialized:"topic_subscriptions"`
}

// shareGroupListOut is used specifically for the list command to exclude partition count
type shareGroupListOut struct {
	Cluster       string `human:"Cluster" serialized:"cluster"`
	ShareGroup    string `human:"Share Group" serialized:"share_group"`
	Coordinator   string `human:"Coordinator" serialized:"coordinator"`
	State         string `human:"State" serialized:"state"`
	ConsumerCount int32  `human:"Consumer Count" serialized:"consumer_count"`
}

func (c *shareCommand) newGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage Kafka share groups.",
	}

	// Only cloud support for now
	cmd.AddCommand(c.newGroupListCommand())
	cmd.AddCommand(c.newGroupDescribeCommand())
	cmd.AddCommand(c.newGroupConsumerCommand())

	return cmd
}

func (c *shareCommand) validGroupArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteShareGroups(cmd, c.AuthenticatedCLICommand)
}
