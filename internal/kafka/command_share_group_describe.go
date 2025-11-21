package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareGroupCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe a Kafka share group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.describe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *shareGroupCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	shareGroup, err := kafkaREST.CloudClient.GetKafkaShareGroup(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&shareGroupOut{
		Cluster:            shareGroup.GetClusterId(),
		ShareGroup:         shareGroup.GetShareGroupId(),
		Coordinator:        getStringBroker(shareGroup.GetCoordinator().Related),
		State:              shareGroup.GetState(),
		ConsumerCount:      shareGroup.GetConsumerCount(),
		PartitionCount:     shareGroup.GetPartitionCount(),
		TopicSubscriptions: getShareGroupTopicNames(shareGroup),
	})
	return table.Print()
}

func getShareGroupTopicNames(shareGroup kafkarestv3.ShareGroupData) []string {
	topicPartitions := shareGroup.GetAssignedTopicPartitions()
	if len(topicPartitions) == 0 {
		return []string{}
	}

	topicSet := make(map[string]bool)
	for _, tp := range topicPartitions {
		topicSet[tp.GetTopicName()] = true
	}

	if len(topicSet) == 0 {
		return []string{}
	}

	topics := make([]string, 0, len(topicSet))
	for topic := range topicSet {
		topics = append(topics, topic)
	}

	return topics
}
