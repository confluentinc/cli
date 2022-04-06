package kafka

import (
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const (
	defaultReplicationFactor = 3
	partitionCount           = "num.partitions"
)

type kafkaTopicCommand struct {
	*hasAPIKeyTopicCommand
	*authenticatedTopicCommand
}

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

func newTopicCommand(cfg *v1.Config, prerunner pcmd.PreRunner, clientID string) *kafkaTopicCommand {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage Kafka topics.",
	}

	c := &kafkaTopicCommand{}

	if cfg.IsCloudLogin() {
		c.hasAPIKeyTopicCommand = &hasAPIKeyTopicCommand{
			HasAPIKeyCLICommand: pcmd.NewHasAPIKeyCLICommand(cmd, prerunner),
			prerunner:           prerunner,
			clientID:            clientID,
		}
		c.hasAPIKeyTopicCommand.init()

		c.authenticatedTopicCommand = &authenticatedTopicCommand{
			AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
			prerunner:                     prerunner,
			clientID:                      clientID,
		}
		c.authenticatedTopicCommand.init()
	} else {
		c.authenticatedTopicCommand = &authenticatedTopicCommand{
			AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
			prerunner:                     prerunner,
			clientID:                      clientID,
		}
		c.authenticatedTopicCommand.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand))
		c.authenticatedTopicCommand.onPremInit()
	}

	return c
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

func (c *hasAPIKeyTopicCommand) init() {
	c.AddCommand(c.newProduceCommand())
	c.AddCommand(c.newConsumeCommand())
}

func (c *authenticatedTopicCommand) init() {
	describeCmd := c.newDescribeCommand()
	updateCmd := c.newUpdateCommand()
	deleteCmd := c.newDeleteCommand()

	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newCreateCommand())
	c.AddCommand(describeCmd)
	c.AddCommand(updateCmd)
	c.AddCommand(deleteCmd)
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
		log.CliLogger.Tracef("Validate topic: found topic " + t.Topic)
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
