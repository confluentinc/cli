package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *authenticatedTopicCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <topic-1> [topic-2] ... [topic-n]",
		Short:             "Delete Kafka topics.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the topics "my_topic" and "my_topic_avro". Use this command carefully as data loss can occur.`,
				Code: "confluent kafka topic delete my_topic\nconfluent kafka topic delete my_topic_avro",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *authenticatedTopicCommand) delete(cmd *cobra.Command, args []string) error {
	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaClusterConfig.ID); err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	if err := c.validateArgs(cmd, kafkaREST, kafkaClusterConfig.ID, args); err != nil {
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
		if r, err := kafkaREST.CloudClient.DeleteKafkaTopic(kafkaClusterConfig.ID, id); err != nil {
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil && restErr.Code == ccloudv2.UnknownTopicOrPartitionErrorCode {
				errs = errors.Join(errs, fmt.Errorf(errors.UnknownTopicErrorMsg, id))
			} else {
				errs = errors.Join(errs, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, r))
			}
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.Topic)

	return errs
}

func (c *authenticatedTopicCommand) validateArgs(cmd *cobra.Command, kafkaREST *pcmd.KafkaREST, clusterId string, args []string) error {
	describeFunc := func(id string) error {
		_, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(clusterId, id)
		return err
	}

	err := deletion.ValidateArgsForDeletion(cmd, args, resource.Topic, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.Topic, "kafka topic"))

	return err
}
