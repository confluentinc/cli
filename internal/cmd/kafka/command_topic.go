package kafka

import (
	"context"
	"fmt"
	"net/http"
	"time"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const (
	badRequestErrorCode              = 40002
	unknownTopicOrPartitionErrorCode = 40403
)

const (
	defaultReplicationFactor = 3
	partitionCount           = "num.partitions"
)

type hasAPIKeyTopicCommand struct {
	*pcmd.HasAPIKeyCLICommand
	prerunner pcmd.PreRunner
	clientID  string
}

type authenticatedTopicCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
	clientID  string
}

type structuredDescribeDisplay struct {
	TopicName string            `json:"topic_name" yaml:"topic_name"`
	Config    map[string]string `json:"config" yaml:"config"`
}

type topicData struct {
	TopicName string            `json:"topic_name" yaml:"topic_name"`
	Config    map[string]string `json:"config" yaml:"config"`
}

func newTopicCommand(cfg *v1.Config, prerunner pcmd.PreRunner, clientID string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage Kafka topics.",
	}

	c := &authenticatedTopicCommand{
		prerunner: prerunner,
		clientID:  clientID,
	}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)

		cmd.AddCommand(newConsumeCommand(prerunner, clientID))
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(newProduceCommand(prerunner, clientID))
		cmd.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newConsumeCommandOnPrem())
		cmd.AddCommand(c.newCreateCommandOnPrem())
		cmd.AddCommand(c.newDeleteCommandOnPrem())
		cmd.AddCommand(c.newDescribeCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
		cmd.AddCommand(c.newProduceCommandOnPrem())
		cmd.AddCommand(c.newUpdateCommandOnPrem())
	}

	return cmd
}

func (c *authenticatedTopicCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTopics()
}

func (c *authenticatedTopicCommand) autocompleteTopics() []string {
	topics, err := c.getTopics()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(topics))
	for i, topic := range topics {
		var description string
		if topic.Internal {
			description = "Internal"
		}
		suggestions[i] = fmt.Sprintf("%s\t%s", topic.Name, description)
	}
	return suggestions
}

// validate that a topic exists before attempting to produce/consume messages
func (c *hasAPIKeyTopicCommand) validateTopic(client *ckafka.AdminClient, topic string, cluster *v1.KafkaClusterConfig) error {
	timeout := 10 * time.Second
	metadata, err := client.GetMetadata(nil, true, int(timeout.Milliseconds()))
	if err != nil {
		if err.Error() == ckafka.ErrTransport.String() {
			err = errors.New("API key may not be provisioned yet")
		}
		return fmt.Errorf("failed to obtain topics from client: %v", err)
	}

	foundTopic := false
	for _, t := range metadata.Topics {
		log.CliLogger.Tracef("Validate topic: found topic %s", t.Topic)
		if topic == t.Topic {
			foundTopic = true // no break so that we see all topics from the above printout
		}
	}
	if !foundTopic {
		log.CliLogger.Trace("validateTopic failed due to topic not being found in the client's topic list")
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsErrorMsg, topic), fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsSuggestions, cluster.ID, cluster.ID, cluster.ID))
	}

	log.CliLogger.Tracef("validateTopic succeeded")
	return nil
}

func (c *authenticatedTopicCommand) getNumPartitions(topicName string) (int, error) {
	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil {
		kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
		if err != nil {
			return 0, err
		}

		partitionsResp, httpResp, err := kafkaREST.CloudClient.ListKafkaPartitions(kafkaClusterConfig.ID, topicName)
		if err != nil && httpResp != nil {
			// Kafka REST is available, but there was an error
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil {
				if restErr.Code == unknownTopicOrPartitionErrorCode {
					return 0, fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
				}
			}
			return 0, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}
		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return 0, errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}

			return len(partitionsResp.Data), nil
		}
	}

	// Fallback to Kafka API
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return 0, err
	}

	topic := &schedv1.TopicSpecification{Name: topicName}
	resp, err := c.Client.Kafka.DescribeTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: false})
	if err != nil {
		return 0, err
	}

	return len(resp.Partitions), nil
}
