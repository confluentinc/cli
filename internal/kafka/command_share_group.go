package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"

	// Import the official SDK types
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
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

func (c *shareCommand) newGroupCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage Kafka share groups.",
	}

	// Only cloud support for now
	cmd.AddCommand(c.newGroupListCommand())
	cmd.AddCommand(c.newGroupDescribeCommand())
	cmd.AddCommand(c.newGroupConsumerCommand(cfg))

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

// Helper function to format unique topics from topic partitions
func formatAssignedTopicPartitions(topicPartitions []kafkarestv3.ShareGroupTopicPartitionData) string {
	if len(topicPartitions) == 0 {
		return "None"
	}

	// Use a map to collect unique topic names
	topicSet := make(map[string]bool)
	for _, tp := range topicPartitions {
		topicSet[tp.GetTopicName()] = true
	}

	if len(topicSet) == 0 {
		return "None"
	}

	// Convert map keys to slice for consistent ordering
	var topics []string
	for topic := range topicSet {
		topics = append(topics, topic)
	}

	return strings.Join(topics, ", ")
}
