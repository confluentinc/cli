package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1), // <topic>
		RunE:  pcmd.NewCLIRunE(c.onPremCreate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a topic named `my_topic` with default options at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic create my_topic --url http://localhost:8082",
			},
			examples.Example{
				Text: "Create a topic named `my_topic_2` with specified configuration parameters.",
				Code: "confluent kafka topic create my_topic_2 --url http://localhost:8082 --config cleanup.policy=compact,compression.type=gzip",
			}),
	}
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	cmd.Flags().Int32("partitions", 6, "Number of topic partitions.")
	cmd.Flags().Int32("replication-factor", 3, "Number of replicas.")
	cmd.Flags().StringSlice("config", nil, "A comma-separated list of topic configuration ('key=value') overrides for the topic being created.")
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")

	return cmd
}

func (c *authenticatedTopicCommand) onPremCreate(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Parse remaining arguments
	numPartitions, err := cmd.Flags().GetInt32("partitions")
	if err != nil {
		return err
	}
	replicationFactor, err := cmd.Flags().GetInt32("replication-factor")
	if err != nil {
		return err
	}
	topicConfigStrings, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	ifNotExists, err := cmd.Flags().GetBool("if-not-exists")
	if err != nil {
		return err
	}

	topicConfigsMap, err := utils.ToMap(topicConfigStrings)
	if err != nil {
		return err
	}
	topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		v2 := v // create a local copy to use pointer
		topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
			Name:  k,
			Value: &v2,
		}
		i++
	}
	// Create new topic
	_, resp, err := restClient.TopicV3Api.CreateKafkaTopic(restContext, clusterId, &kafkarestv3.CreateKafkaTopicOpts{
		CreateTopicRequestData: optional.NewInterface(kafkarestv3.CreateTopicRequestData{
			TopicName:         topicName,
			PartitionsCount:   numPartitions,
			ReplicationFactor: replicationFactor,
			Configs:           topicConfigs,
		}),
	})
	if err != nil {
		// catch topic exists error
		if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
			var decodedError kafkaRestV3Error
			err2 := json.Unmarshal(openAPIError.Body(), &decodedError)
			if err2 != nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			if decodedError.Message == fmt.Sprintf("Topic '%s' already exists.", topicName) {
				if !ifNotExists {
					return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicExistsOnPremErrorMsg, topicName), errors.TopicExistsOnPremSuggestions)
				} // ignore error if ifNotExists flag is set
				return nil
			}
		}
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	// no error if topic is created successfully.
	utils.Printf(cmd, errors.CreatedTopicMsg, topicName)
	return nil
}
