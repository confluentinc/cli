package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/serdes"
)

// EOF Unicode encoding for Ctrl+D (^D) character
const EOF = "\u0004"

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

		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newCreateCommandOnPrem())
		cmd.AddCommand(c.newDeleteCommandOnPrem())
		cmd.AddCommand(c.newDescribeCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
		cmd.AddCommand(c.newUpdateCommandOnPrem())
	}

	cmd.AddCommand(c.newConfigurationCommand(cfg))
	cmd.AddCommand(c.newConsumeCommand())
	cmd.AddCommand(c.newProduceCommand())

	return cmd
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *command) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTopics(cmd)
}

func (c *command) autocompleteTopics(cmd *cobra.Command) []string {
	topics, err := c.getTopics(cmd)
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
func (c *command) validateTopic(client *ckgo.AdminClient, topic string, cluster *config.KafkaClusterConfig) error {
	timeout := 10 * time.Second
	metadata, err := client.GetMetadata(nil, true, int(timeout.Milliseconds()))
	if err != nil {
		if err.Error() == ckgo.ErrTransport.String() {
			err = fmt.Errorf("API key may not be provisioned yet")
		}
		return fmt.Errorf("failed to obtain topics from client: %w", err)
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
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.TopicDoesNotExistOrMissingPermissionsErrorMsg, topic),
			fmt.Sprintf(errors.TopicDoesNotExistOrMissingPermissionsSuggestions, cluster.ID),
		)
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
		return fmt.Errorf(errors.KafkaRestProvisioningErrorMsg, lkc)
	}
	return nil
}

func (c *command) prepareAnonymousContext(cmd *cobra.Command) error {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return err
	}

	platform := &config.Platform{
		Server: bootstrap,
		Name:   strings.TrimPrefix(bootstrap, "https://"),
	}

	kafkaClusterCfg := &config.KafkaClusterConfig{
		ID:        "anonymous-id",
		Bootstrap: bootstrap,
		APIKeys:   map[string]*config.APIKeyPair{},
	}
	kafkaClusters := map[string]*config.KafkaClusterConfig{kafkaClusterCfg.ID: kafkaClusterCfg}

	c.Context = &config.Context{Platform: platform}
	c.Context.KafkaClusterContext = &config.KafkaClusterContext{
		EnvContext:          false,
		ActiveKafkaCluster:  kafkaClusterCfg.ID,
		KafkaClusterConfigs: kafkaClusters,
		Context:             c.Context,
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
			fmt.Sprintf(`no secret for API key "%s" of resource "%s" passed via flag or stored in local CLI state`, apiKey, cluster.ID),
			fmt.Sprintf("Pass the API secret with flag `--api-secret` or store with `confluent api-key store %s --resource %s`.", apiKey, cluster.ID),
		)
	}

	return nil
}

func (c *command) produceToTopic(cmd *cobra.Command, keyMetaInfo []byte, valueMetaInfo []byte, topic string, keySerializer serdes.SerializationProvider, valueSerializer serdes.SerializationProvider, producer *ckgo.Producer, isOnPrem bool) error {
	keys := "Ctrl-C or Ctrl-D"
	if runtime.GOOS == "windows" {
		keys = "Ctrl-C"
	}
	output.ErrPrintf(false, "Starting Kafka Producer. Use %s to exit.\n", keys)
	if isOnPrem {
		go func(eventsChan chan ckgo.Event) {
			for ev := range eventsChan {
				oart, _ := ev.(ckgo.OAuthBearerTokenRefresh)
				err := c.refreshOAuthBearerToken(cmd, producer, oart)
				if err != nil {
					return
				}
			}
		}(producer.Events())
	}
	var scanErr error
	input, scan := PrepareInputChannel(&scanErr)

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		input <- EOF
	}()
	// Prime reader
	go scan()

	deliveryChan := make(chan ckgo.Event)
	for data := range input {
		if data == "" {
			if scanErr != nil {
				break
			}
			go scan()
			continue
		} else if data == EOF {
			break
		}

		message, err := GetProduceMessage(cmd, keyMetaInfo, valueMetaInfo, topic, data, keySerializer, valueSerializer)
		if err != nil {
			return err
		}
		if err := producer.Produce(message, deliveryChan); err != nil {
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topic)
			if isProduceToCompactedTopicError {
				scanErr = err
				break
			}
			output.ErrPrintf(false, errors.FailedToProduceErrorMsg, message.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckgo.Message)             // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			output.ErrPrintf(false, errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	return scanErr
}

func createTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	err := os.MkdirAll(dir, 0755)
	return dir, err
}
