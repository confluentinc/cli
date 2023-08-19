package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <topic-1> [topic-2] ... [topic-n]",
		Short:             "Delete one or more Kafka topics.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
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

func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaREST.GetClusterId()); err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.Topic, args[0]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if r, err := kafkaREST.CloudClient.DeleteKafkaTopic(id); err != nil {
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil && restErr.Code == ccloudv2.UnknownTopicOrPartitionErrorCode {
				return fmt.Errorf(errors.UnknownTopicErrorMsg, id)
			} else {
				return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, r)
			}
		}
		return nil
	}

	_, err = deletion.Delete(args, deleteFunc, resource.Topic)
	return err
}
