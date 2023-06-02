package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *authenticatedTopicCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a topic named "my_topic" with default options for the current specified cluster (providing embedded Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic create my_topic --url http://localhost:8090/kafka",
			},
			examples.Example{
				Text: `Create a topic named "my_topic" with default options at specified cluster (providing Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic create my_topic --url http://localhost:8082",
			},
			examples.Example{
				Text: `Create a topic named "my_topic_2" with specified configuration parameters.`,
				Code: "confluent kafka topic create my_topic_2 --url http://localhost:8082 --config cleanup.policy=compact,compression.type=gzip",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	cmd.Flags().Uint32("partitions", 0, "Number of topic partitions.")
	cmd.Flags().Uint32("replication-factor", 0, "Number of replicas.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of topic configuration ("key=value") overrides for the topic being created.`)
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")

	return cmd
}

func (c *authenticatedTopicCommand) createOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	return CreateTopic(cmd, restClient, restContext, topicName, clusterId)
}

func CreateTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	partitions, err := cmd.Flags().GetUint32("partitions")
	if err != nil {
		return err
	}

	replicationFactor, err := cmd.Flags().GetUint32("replication-factor")
	if err != nil {
		return err
	}

	ifNotExists, err := cmd.Flags().GetBool("if-not-exists")
	if err != nil {
		return err
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.ConfigFlagToMap(configs)
	if err != nil {
		return err
	}

	topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(configMap))
	i := 0
	for k, v := range configMap {
		v2 := v // create a local copy to use pointer
		topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
			Name:  k,
			Value: &v2,
		}
		i++
	}

	data := kafkarestv3.CreateTopicRequestData{
		TopicName: topicName,
		Configs:   topicConfigs,
	}

	if cmd.Flags().Changed("partitions") {
		data.PartitionsCount = int32(partitions)
	}

	if cmd.Flags().Changed("replication-factor") {
		data.ReplicationFactor = int32(replicationFactor)
	}

	opts := &kafkarestv3.CreateKafkaTopicOpts{CreateTopicRequestData: optional.NewInterface(data)}

	// Create new topic
	if _, resp, err := restClient.TopicV3Api.CreateKafkaTopic(restContext, clusterId, opts); err != nil {
		// catch topic exists error
		if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
			var decodedError kafkarest.V3Error
			if err := json.Unmarshal(openAPIError.Body(), &decodedError); err != nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			if decodedError.Message == fmt.Sprintf("Topic '%s' already exists.", topicName) {
				if !ifNotExists {
					return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicExistsOnPremErrorMsg, topicName), errors.TopicExistsOnPremSuggestions)
				} // ignore error if ifNotExists flag is set
				return nil
			}
		}
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	output.Printf(errors.CreatedResourceMsg, resource.Topic, topicName)
	return nil
}
