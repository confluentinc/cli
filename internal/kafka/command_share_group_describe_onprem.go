package kafka

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareGroupCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <group>",
		Short: "Describe a Kafka share group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *shareGroupCommand) describeOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	shareGroup, resp, err := restClient.ShareGroupV3Api.GetKafkaShareGroup(restContext, clusterId, args[0])
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	table.Add(&shareGroupOut{
		Cluster:            shareGroup.ClusterId,
		ShareGroup:         shareGroup.ShareGroupId,
		Coordinator:        getStringBroker(shareGroup.Coordinator.Related),
		State:              shareGroup.State,
		ConsumerCount:      shareGroup.ConsumerCount,
		PartitionCount:     shareGroup.PartitionCount,
		TopicSubscriptions: getShareGroupTopicNamesOnPrem(shareGroup),
	})
	return table.Print()
}

func getShareGroupTopicNamesOnPrem(shareGroup kafkarestv3.ShareGroupData) []string {
	if len(shareGroup.AssignedTopicPartitions) == 0 {
		return []string{}
	}

	topicSet := make(map[string]bool)
	for _, tp := range shareGroup.AssignedTopicPartitions {
		topicSet[tp.TopicName] = true
	}

	if len(topicSet) == 0 {
		return []string{}
	}

	topics := make([]string, 0, len(topicSet))
	for topic := range topicSet {
		topics = append(topics, topic)
	}

	sort.Strings(topics)
	return topics
}
