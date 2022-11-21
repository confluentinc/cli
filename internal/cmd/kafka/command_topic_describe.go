package kafka

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
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

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaClusterConfig.ID); err != nil {
		return err
	}

	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil {
		// Get topic config
		configsResp, httpResp, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(kafkaClusterConfig.ID, topicName)
		if err != nil && httpResp != nil {
			// Kafka REST is available, but there was an error
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil {
				if restErr.Code == unknownTopicOrPartitionErrorCode {
					return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
				}
			}
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}

			// Kafka REST is available and there was no error. Fetch partition and config information.
			configs := make(map[string]string)

			for _, config := range configsResp.Data {
				configs[config.Name] = config.GetValue()
			}
			numPartitions, err := c.getNumPartitions(topicName)
			if err != nil {
				return err
			}
			configs[partitionCount] = strconv.Itoa(numPartitions)

			if output.GetFormat(cmd).IsSerialized() {
				return output.SerializedOutput(cmd, configs)
			}

			list := output.NewList(cmd)
			for name, value := range configs {
				list.Add(&configOut{
					Name:  name,
					Value: value,
				})
			}
			list.Filter([]string{"Name", "Value"})
			return list.Print()
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topicName}}
	resp, err := c.PrivateClient.Kafka.DescribeTopic(context.Background(), cluster, topic)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd).IsSerialized() {
		out := make(map[string]string)
		for _, entry := range resp.Config {
			out[entry.Name] = entry.Value
		}
		out[partitionCount] = strconv.Itoa(len(resp.Partitions))
		return output.SerializedOutput(cmd, out)
	}

	list := output.NewList(cmd)
	for _, config := range resp.Config {
		list.Add(&configOut{
			Name:  config.Name,
			Value: config.Value,
		})
	}
	list.Add(&configOut{
		Name:  partitionCount,
		Value: strconv.Itoa(len(resp.Partitions)),
	})
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
