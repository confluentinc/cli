package kafka

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *authenticatedTopicCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
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

func (c *authenticatedTopicCommand) deleteOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	return DeleteTopicWithRestClient(cmd, restClient, restContext, topicName, clusterId)
}

func DeleteTopicWithRestClient(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	if _, resp, err := restClient.TopicV3Api.GetKafkaTopic(restContext, clusterId, topicName); err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.Topic, topicName, topicName)
	if _, err := form.ConfirmDeletion(cmd, promptMsg, topicName); err != nil {
		return err
	}

	if resp, err := restClient.TopicV3Api.DeleteKafkaTopic(restContext, clusterId, topicName); err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	output.Printf(errors.DeletedResourceMsg, resource.Topic, topicName)
	return nil
}
