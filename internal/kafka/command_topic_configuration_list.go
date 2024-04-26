package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <topic>",
		Short:             "List Kafka topic configurations.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.configurationList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List configurations for topic "my-topic".`,
				Code: "confluent kafka topic configuration list my-topic",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) configurationList(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaREST.GetClusterId()); err != nil {
		return err
	}

	configs, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(topicName)
	if err != nil {
		return err
	}

	topic, httpResp, err := kafkaREST.CloudClient.GetKafkaTopic(topicName)
	if err != nil {
		if restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err); parseErr == nil && restErr.Code == ccloudv2.UnknownTopicOrPartitionErrorCode {
			return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
		}
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	list := output.NewList(cmd)
	for _, config := range configs {
		list.Add(&topicConfigurationOut{
			Name:     config.GetName(),
			Value:    config.GetValue(),
			ReadOnly: config.GetIsReadOnly(),
		})
	}
	list.Add(&topicConfigurationOut{
		Name:     numPartitionsKey,
		Value:    fmt.Sprint(topic.PartitionsCount),
		ReadOnly: false,
	})
	return list.Print()
}
