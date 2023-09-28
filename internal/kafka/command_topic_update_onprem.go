package kafka

import (
	"context"
	"sort"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
)

func (c *command) newUpdateCommandOnPrem() *cobra.Command {
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
	pcmd.AddConfigFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) updateOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return UpdateTopic(cmd, restClient, restContext, topicName, clusterId)
}

func UpdateTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.GetMap(configs)
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

	output.Printf("Updated the following configuration values for topic \"%s\":\n", topicName)

	list := output.NewList(cmd)
	for _, config := range data {
		list.Add(&broker.ConfigOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
