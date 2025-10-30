package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type shareGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type shareGroupOut struct {
	Cluster            string   `human:"Cluster" serialized:"cluster"`
	ShareGroup         string   `human:"Share Group" serialized:"share_group"`
	Coordinator        string   `human:"Coordinator" serialized:"coordinator"`
	State              string   `human:"State" serialized:"state"`
	ConsumerCount      int32    `human:"Consumer Count" serialized:"consumer_count"`
	PartitionCount     int32    `human:"Partition Count" serialized:"partition_count"`
	TopicSubscriptions []string `human:"Topic Subscriptions,omitempty" serialized:"topic_subscriptions,omitempty"`
}

func newShareGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "share-group",
		Short:       "Manage Kafka share groups.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &shareGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newConsumerCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func (c *shareGroupCommand) validGroupArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteShareGroups(cmd, c.AuthenticatedCLICommand)
}
