package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				// on-prem examples are ccloud examples + "of a specified cluster (providing embedded Kafka REST Proxy endpoint)."
				Text: `List all topics for a specified cluster (providing Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic list --url http://localhost:8090/kafka",
			},

			examples.Example{
				// on-prem examples are ccloud examples + "of a specified cluster (providing Kafka REST Proxy endpoint)."
				Text: "List all topics for a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listOnPrem(cmd *cobra.Command, _ []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return ListTopics(cmd, restClient, restContext, clusterId)
}

func ListTopics(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	topics, resp, err := restClient.TopicV3Api.ListKafkaTopics(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, topic := range topics.Data {
		list.Add(&topicOut{
			Name:              topic.TopicName,
			IsInternal:        topic.IsInternal,
			ReplicationFactor: topic.ReplicationFactor,
			PartitionCount:    topic.PartitionsCount,
		})
	}

	return list.Print()
}
