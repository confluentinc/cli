package kafka

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "consume <topic>",
		Short:             "Consume messages from a Kafka topic.",
		Long:              "Consume messages from a Kafka topic.\n\nTruncated message headers will be printed if they exist.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.consume,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume items from the "my_topic" topic and press "Ctrl-C" to exit.`,
				Code: "confluent kafka topic consume -b my_topic",
			},
		),
	}

	cmd.Flags().String("group", "confluent_cli_consumer_<randomly-generated-id>", "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().Int64("offset", 0, "The offset from the beginning to consume from.")
	cmd.Flags().Int32("partition", -1, "The partition to consume from.")
	pcmd.AddKeyFormatFlag(cmd)
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("full-header", false, "Print complete content of message headers.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().Bool("timestamp", false, "Print message timestamp in milliseconds.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	pcmd.AddConsumerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-context", "", "The Schema Registry context under which to look up schema ID.")
	cmd.Flags().String("schema-registry-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().String("schema-registry-api-key", "", "Schema registry API key.")
	cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API key secret.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("config", "config-file")
	cmd.MarkFlagsMutuallyExclusive("from-beginning", "offset")

	return cmd
}

func (c *command) consume(cmd *cobra.Command, args []string) error {
	topic := args[0]

	cluster, err := c.Config.Context().GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := addApiKeyToCluster(cmd, cluster); err != nil {
		return err
	}

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("group") {
		group = fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	}

	printKey, err := cmd.Flags().GetBool("print-key")
	if err != nil {
		return err
	}

	fullHeader, err := cmd.Flags().GetBool("full-header")
	if err != nil {
		return err
	}

	timestamp, err := cmd.Flags().GetBool("timestamp")
	if err != nil {
		return err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return err
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
		return fmt.Errorf(errors.FailedToCreateConsumerErrorMsg, err)
	}
	log.CliLogger.Trace("Create consumer succeeded")

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	if err := c.validateTopic(adminClient, topic, cluster); err != nil {
		return err
	}

	offset, err := GetOffsetWithFallback(cmd)
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}
	partitionFilter := PartitionFilter{
		Changed: cmd.Flags().Changed("partition"),
		Index:   partition,
	}

	rebalanceCallback := GetRebalanceCallback(offset, partitionFilter)
	if err := consumer.Subscribe(topic, rebalanceCallback); err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingConsumerMsg)

	keyFormat, err := cmd.Flags().GetString("key-format")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	var srClient *srsdk.APIClient
	var ctx context.Context
	if valueFormat != "string" {
		schemaRegistryApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
		if err != nil {
			return err
		}
		schemaRegistryApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
		if err != nil {
			return err
		}
		// Only initialize client and context when schema is specified.
		srClient, ctx, err = sr.GetSchemaRegistryClientWithApiKey(cmd, c.Config, c.Version, schemaRegistryApiKey, schemaRegistryApiSecret)
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
	schemaRegistryContext, err := cmd.Flags().GetString("schema-registry-context")
	if err != nil {
		return err
	}
	if schemaRegistryContext != "" {
		subject = schemaRegistryContext
	}

	groupHandler := &GroupHandler{
		SrClient:    srClient,
		Ctx:         ctx,
		KeyFormat:   keyFormat,
		ValueFormat: valueFormat,
		Out:         cmd.OutOrStdout(),
		Subject:     subject,
		Properties: ConsumerProperties{
			PrintKey:   printKey,
			FullHeader: fullHeader,
			Timestamp:  timestamp,
			Delimiter:  delimiter,
			SchemaPath: dir,
		},
	}
	return RunConsumer(consumer, groupHandler)
}
