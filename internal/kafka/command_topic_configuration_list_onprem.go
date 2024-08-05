package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newConfigurationListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationListOnPrem,
		Short: "List Kafka topic configurations.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List configurations for topic "my-topic" for the specified cluster (providing embedded Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic configuration list my-topic --url http://localhost:8090/kafka",
			},

			examples.Example{
				Text: `List configurations for topic "my-topic" for the specified cluster (providing Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic configuration list my-topic --url http://localhost:8082",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) configurationListOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return ListConfigurations(cmd, restClient, restContext, topicName, clusterId)
}

func ListConfigurations(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	configs, resp, err := restClient.ConfigsV3Api.ListKafkaTopicConfigs(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	} else if configs.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}

	list := output.NewList(cmd)
	for _, config := range configs.Data {
		out := &broker.ConfigOut{Name: config.Name}
		if config.Value != nil {
			out.Value = *config.Value
		}
		list.Add(out)
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
