package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *authenticatedTopicCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <topic>",
		Short:             "Describe a Kafka topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my_topic" topic.`,
				Code: "confluent kafka topic describe my_topic",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) describe(cmd *cobra.Command, args []string) error {
	topicName := args[0]

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

	configsResp, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(kafkaClusterConfig.ID, topicName)
	if err != nil {
		return err
	}

	topicData, httpResp, err := kafkaREST.CloudClient.GetKafkaTopic(kafkaClusterConfig.ID, topicName)
	if err != nil {
		if restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err); parseErr == nil && restErr.Code == ccloudv2.UnknownTopicOrPartitionErrorCode {
			return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
		}
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	list := output.NewList(cmd)
	for _, config := range configsResp.Data {
		list.Add(&topicConfigurationOut{
			Name:     config.GetName(),
			Value:    config.GetValue(),
			ReadOnly: config.GetIsReadOnly(),
		})
	}
	list.Add(&topicConfigurationOut{
		Name:     numPartitionsKey,
		Value:    fmt.Sprint(topicData.PartitionsCount),
		ReadOnly: false,
	})
	return list.Print()
}
