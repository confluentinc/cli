package kafka

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *authenticatedTopicCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <topic-1> [topic-2] ... [topic-n]",
		Short: "Delete Kafka topics.",
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

	if err := c.validateArgsOnPrem(cmd, restClient, restContext, clusterId, args); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.Topic, args[0], args[0]); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.Topic, args); err != nil || !ok {
			return err
		}
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if r, err := restClient.TopicV3Api.DeleteKafkaTopic(restContext, clusterId, id); err != nil {
			errs = errors.Join(errs, kafkarest.NewError(restClient.GetConfig().BasePath, err, r))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.Topic)

	return errs
}

func (c *authenticatedTopicCommand) validateArgsOnPrem(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, args []string) error {
	describeFunc := func(id string) error {
		_, _, err := restClient.TopicV3Api.GetKafkaTopic(restContext, clusterId, id)
		return err
	}

	err := deletion.ValidateArgsForDeletion(cmd, args, resource.Topic, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.Topic, "kafka topic"))

	return err
}
