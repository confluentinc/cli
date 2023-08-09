package kafka

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const numPartitionsKey = "num.partitions"

type command struct {
	*pcmd.AuthenticatedCLICommand
	clientID string
}

func newTopicCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage Kafka topics.",
	}

	c := &command{clientID: cfg.Version.ClientID}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newConsumeCommand())
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newProduceCommand())
		cmd.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
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

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTopics()
}

func (c *command) autocompleteTopics() []string {
	topics, err := c.getTopics()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(topics))
	for i, topic := range topics {
		var description string
		if topic.GetIsInternal() {
			description = "Internal"
		}
		suggestions[i] = fmt.Sprintf("%s\t%s", topic.GetTopicName(), description)
	}
	return suggestions
}

// validate that a topic exists before attempting to produce/consume messages
func (c *command) validateTopic(client *ckafka.AdminClient, topic string, cluster *config.KafkaClusterConfig) error {
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
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicDoesNotExistOrMissingPermissionsErrorMsg, topic), fmt.Sprintf(errors.TopicDoesNotExistOrMissingPermissionsSuggestions, cluster.ID, cluster.ID, cluster.ID))
	}

	log.CliLogger.Tracef("validateTopic succeeded")
	return nil
}

func (c *command) provisioningClusterCheck(lkc string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, environmentId)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc, httpResp)
	}
	if cluster.Status.Phase == ccloudv2.StatusProvisioning {
		return errors.Errorf(errors.KafkaRestProvisioningErrorMsg, lkc)
	}
	return nil
}

func addApiKeyToCluster(cmd *cobra.Command, cluster *config.KafkaClusterConfig) error {
	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}

	if apiKey != "" {
		apiSecret, err := cmd.Flags().GetString("api-secret")
		if err != nil {
			return err
		}

		cluster.APIKey = apiKey
		cluster.APIKeys[cluster.APIKey] = &config.APIKeyPair{
			Key:    apiKey,
			Secret: apiSecret,
		}
	}

	if cluster.APIKey == "" {
		return &errors.UnspecifiedAPIKeyError{ClusterID: cluster.ID}
	}

	if pair, ok := cluster.APIKeys[cluster.APIKey]; !ok || pair.Secret == "" {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.NoAPISecretStoredOrPassedErrorMsg, apiKey, cluster.ID),
			fmt.Sprintf(errors.NoAPISecretStoredOrPassedSuggestions, apiKey, cluster.ID))
	}

	return nil
}
