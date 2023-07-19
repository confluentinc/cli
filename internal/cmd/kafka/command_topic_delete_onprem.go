package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *authenticatedTopicCommand) newDeleteCommandOnPrem() *cobra.Command {
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

func (c *authenticatedTopicCommand) deleteOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	return DeleteTopic(cmd, restClient, restContext, args, clusterId)
}

func DeleteTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, args []string, clusterId string) error {
	if confirm, err := confirmDeletionOnPrem(cmd, restClient, restContext, clusterId, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := restClient.TopicV3Api.DeleteKafkaTopic(restContext, clusterId, id); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, r)
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	resource.PrintDeleteSuccessMsg(deleted, resource.Topic)

	return err
}

func confirmDeletionOnPrem(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, _, err := restClient.TopicV3Api.GetKafkaTopic(restContext, clusterId, id)
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.Topic, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.Topic, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.Topic, args[0], args[0]), args[0]); err != nil {
		return false, err
	}

	return true, nil
}
