package kafka

import (
	"context"
	"fmt"
	"os"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func newConsumeCommand(prerunner pcmd.PreRunner, clientId string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "consume <topic>",
		Short:       "Consume messages from a Kafka topic.",
		Long:        "Consume messages from a Kafka topic.\n\nTruncated message headers will be printed if they exist.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume items from the "my_topic" topic and press "Ctrl+C" to exit.`,
				Code: "confluent kafka topic consume -b my_topic",
			},
		),
	}

	c := &hasAPIKeyTopicCommand{
		HasAPIKeyCLICommand: pcmd.NewHasAPIKeyCLICommand(cmd, prerunner),
		prerunner:           prerunner,
		clientID:            clientId,
	}
	cmd.RunE = c.consume

	cmd.Flags().String("group", fmt.Sprintf("confluent_cli_consumer_%s", uuid.New()), "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().Int64("offset", 0, "The offset from the beginning to consume from.")
	cmd.Flags().Int32("partition", -1, "The partition to consume from.")
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("full-header", false, "Print complete content of message headers.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in json or avro format) for the consumer client.")
	cmd.Flags().String("context-name", "", "The Schema Registry context under which to lookup schema ID.")
	cmd.Flags().String("sr-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().String("sr-api-key", "", "Schema registry API key.")
	cmd.Flags().String("sr-api-secret", "", "Schema registry API key secret.")
	cmd.Flags().String("api-key", "", "API key.")
	cmd.Flags().String("api-secret", "", "API key secret.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	cmd.Flags().String("environment", "", "Environment ID.")

	return cmd
}

func (c *hasAPIKeyTopicCommand) consume(cmd *cobra.Command, args []string) error {
	topic := args[0]

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	cluster, err := c.Config.Context().GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	printKey, err := cmd.Flags().GetBool("print-key")
	if err != nil {
		return err
	}

	fullHeader, err := cmd.Flags().GetBool("full-header")
	if err != nil {
		return err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("config-file") && cmd.Flags().Changed("config") {
		return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "config-file", "config")
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	consumer, err := newConsumer(group, cluster, c.clientID, configFile, config)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateConsumerMsg, err)
	}
	log.CliLogger.Trace("Create consumer succeeded")

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientMsg, err)
	}
	defer adminClient.Close()

	err = c.validateTopic(adminClient, topic, cluster)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("from-beginning") && cmd.Flags().Changed("offset") {
		return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "from-beginning", "offset")
	}

	offset, err := getOffsetWithFallback(cmd)
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}
	partitionFilter := partitionFilter{
		changed: cmd.Flags().Changed("partition"),
		index:   partition,
	}

	rebalanceCallback := getRebalanceCallback(cmd, offset, partitionFilter)
	err = consumer.Subscribe(topic, rebalanceCallback)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingConsumerMsg)

	var srClient *srsdk.APIClient
	var ctx context.Context
	if valueFormat != "string" {
		srAPIKey, err := cmd.Flags().GetString("sr-api-key")
		if err != nil {
			return err
		}
		srAPISecret, err := cmd.Flags().GetString("sr-api-secret")
		if err != nil {
			return err
		}
		// Only initialize client and context when schema is specified.
		srClient, ctx, err = sr.GetAPIClientWithAPIKey(cmd, nil, c.Config, c.Version, srAPIKey, srAPISecret)
		if err != nil {
			if err.Error() == errors.NotLoggedInErrorMsg {
				return new(errors.SRNotAuthenticatedError)
			} else {
				return err
			}
		}
	}

	dir, err := sr.CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	subject := topicNameStrategy(topic)
	contextName, err := cmd.Flags().GetString("context-name")
	if err != nil {
		return err
	}
	if contextName != "" {
		subject = contextName
	}

	groupHandler := &GroupHandler{
		SrClient:   srClient,
		Ctx:        ctx,
		Format:     valueFormat,
		Out:        cmd.OutOrStdout(),
		Subject:    subject,
		Properties: ConsumerProperties{PrintKey: printKey, FullHeader: fullHeader, Delimiter: delimiter, SchemaPath: dir},
	}
	return runConsumer(cmd, consumer, groupHandler)
}
