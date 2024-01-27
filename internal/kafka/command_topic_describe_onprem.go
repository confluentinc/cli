package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeOnPrem,
		Short: "Describe a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my_topic" topic for the specified cluster (providing embedded Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic describe my_topic --url http://localhost:8090/kafka",
			},

			examples.Example{
				Text: `Describe the "my_topic" topic for the specified cluster (providing Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic describe my_topic --url http://localhost:8082",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return DescribeTopic(cmd, restClient, restContext, topicName, clusterId)
}

func DescribeTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	topic, resp, err := restClient.TopicV3Api.GetKafkaTopic(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	table.Add(&topicOut{
		Name:              topic.TopicName,
		IsInternal:        topic.IsInternal,
		ReplicationFactor: topic.ReplicationFactor,
		PartitionCount:    topic.PartitionsCount,
	})
	return table.Print()
}
