package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <topic-1> [topic-2] ... [topic-n]",
		Short: "Delete one or more Kafka topics.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.deleteOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the topic "my_topic" for the specified cluster (providing embedded Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.`,
				Code: "confluent kafka topic delete my_topic --url http://localhost:8090/kafka",
			},
			examples.Example{
				Text: `Delete the topic "my_topic" for the specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.`,
				Code: "confluent kafka topic delete my_topic --url http://localhost:8082",
			}),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) deleteOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return DeleteTopic(cmd, restClient, restContext, args, clusterId)
}

func DeleteTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, args []string, clusterId string) error {
	existenceFunc := func(id string) bool {
		_, _, err := restClient.TopicV3Api.GetKafkaTopic(restContext, clusterId, id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.Topic, args[0]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if httpResp, err := restClient.TopicV3Api.DeleteKafkaTopic(restContext, clusterId, id); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
		}
		return nil
	}

	_, err := deletion.Delete(args, deleteFunc, resource.Topic)
	return err
}
