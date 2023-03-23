package kafka

import (
	"context"
	"sort"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
)

func (c *authenticatedTopicCommand) newUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.updateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Modify the "my_topic" topic for the specified cluster (providing embedded Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).`,
				Code: "confluent kafka topic update my_topic --url http://localhost:8082 --config retention.ms=259200000",
			},
			examples.Example{
				Text: `Modify the "my_topic" topic for the specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).`,
				Code: "confluent kafka topic update my_topic --url http://localhost:8082 --config retention.ms=259200000",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of topics configuration ("key=value") overrides for the topic being created.`)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) updateOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	return UpdateTopicWithRESTClient(cmd, restClient, restContext, topicName, clusterId)
}

func UpdateTopicWithRESTClient(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	// Update Config
	configs, err := cmd.Flags().GetStringSlice("config") // handle config parsing errors
	if err != nil {
		return err
	}
	configMap, err := properties.ConfigFlagToMap(configs)
	if err != nil {
		return err
	}

	data := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configMap))
	i := 0
	for key, val := range configMap {
		v := val
		data[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: &v,
		}
		i++
	}

	opts := &kafkarestv3.UpdateKafkaTopicConfigBatchOpts{
		AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: data}),
	}
	resp, err := restClient.ConfigsV3Api.UpdateKafkaTopicConfigBatch(restContext, clusterId, topicName, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	if output.GetFormat(cmd).IsSerialized() {
		sort.Slice(data, func(i, j int) bool {
			return data[i].Name < data[j].Name
		})
		return output.SerializedOutput(cmd, data)
	}

	output.Printf(errors.UpdateTopicConfigMsg, topicName)

	list := output.NewList(cmd)
	for _, config := range data {
		list.Add(&configOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
