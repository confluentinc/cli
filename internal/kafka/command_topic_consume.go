package kafka

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/schemaregistry"
	"github.com/confluentinc/cli/v3/pkg/serdes"
)

func (c *command) newConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "consume <topic>",
		Short:             "Consume messages from a Kafka topic.",
		Long:              "Consume messages from a Kafka topic.\n\nTruncated message headers will be printed if they exist.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.consume,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume items from topic "my-topic" and press "Ctrl-C" to exit.`,
				Code: "confluent kafka topic consume my-topic --from-beginning",
			},
			examples.Example{
				Text: `Consume from a cloud Kafka topic named "my_topic" without logging in to Confluent Cloud.`,
				Code: "confluent kafka topic consume my_topic --api-key 0000000000000000 --api-secret <API_SECRET> --bootstrap SASL_SSL://pkc-12345.us-west-2.aws.confluent.cloud:9092 --value-format avro --schema-registry-endpoint https://psrc-12345.us-west-2.aws.confluent.cloud --schema-registry-api-key 0000000000000000 --schema-registry-api-secret <SCHEMA_REGISTRY_API_SECRET>",
			},
		),
	}

	cmd.Flags().String("bootstrap", "", `Kafka cluster endpoint (Confluent Cloud) or a comma-separated list of broker hosts, each formatted as "host" or "host:port" (Confluent Platform).`)
	cmd.Flags().String("group", "confluent_cli_consumer_<randomly-generated-id>", "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().Int64("offset", 0, "The offset from the beginning to consume from.")
	cmd.Flags().Int32("partition", -1, "The partition to consume from.")
	pcmd.AddKeyFormatFlag(cmd)
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("print-offset", false, "Print partition number and offset of the message.")
	cmd.Flags().Bool("full-header", false, "Print complete content of message headers.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().Bool("timestamp", false, "Print message timestamp in milliseconds.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	pcmd.AddConsumerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-endpoint", "", "Endpoint for Schema Registry cluster.")

	// cloud-only flags
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	cmd.Flags().String("schema-registry-context", "", "The Schema Registry context under which to look up schema ID.")
	cmd.Flags().String("schema-registry-api-key", "", "Schema registry API key.")
	cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API secret.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	// on-prem only flags
	cmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet())
	pcmd.AddProtocolFlag(cmd)
	pcmd.AddMechanismFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("config", "config-file")
	cmd.MarkFlagsMutuallyExclusive("from-beginning", "offset")

	return cmd
}

func (c *command) consume(cmd *cobra.Command, args []string) error {
	if c.Context.GetConfig().IsCloudLogin() {
		return c.consumeCloud(cmd, args)
	}

	if !cmd.Flags().Changed("bootstrap") { // Required if the user isn't logged into Confluent Cloud
		return fmt.Errorf(errors.RequiredFlagNotSetErrorMsg, "bootstrap")
	}

	if c.Context.GetState() == nil {
		bootstrap, err := cmd.Flags().GetString("bootstrap")
		if err != nil {
			return err
		}

		if strings.Contains(bootstrap, "confluent.cloud") {
			if err := c.prepareAnonymousContext(cmd); err != nil {
				return err
			}

			return c.consumeCloud(cmd, args)
		}
	}

	return c.consumeOnPrem(cmd, args)
}

func (c *command) consumeCloud(cmd *cobra.Command, args []string) error {
	topic := args[0]

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
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
	var consumeFromGroupOffset bool
	if cmd.Flags().Changed("group") {
		consumeFromGroupOffset = true
	} else {
		group = fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	}

	printKey, err := cmd.Flags().GetBool("print-key")
	if err != nil {
		return err
	}

	printOffset, err := cmd.Flags().GetBool("print-offset")
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
	if consumeFromGroupOffset && !cmd.Flags().Changed("from-beginning") && !cmd.Flags().Changed("offset") {
		rebalanceCallback = nil
	}
	if err := consumer.Subscribe(topic, rebalanceCallback); err != nil {
		return err
	}

	output.ErrPrintln(c.Config.EnableColor, errors.StartingConsumerMsg)

	keyFormat, err := cmd.Flags().GetString("key-format")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	var srClient *schemaregistry.Client
	if slices.Contains(serdes.SchemaBasedFormats, valueFormat) || slices.Contains(serdes.SchemaBasedFormats, keyFormat) {
		// Only initialize client and context when schema is specified.
		srClient, err = c.GetSchemaRegistryClient(cmd)
		if err != nil {
			if err.Error() == errors.NotLoggedInErrorMsg {
				return new(errors.SRNotAuthenticatedError)
			} else {
				return err
			}
		}
	}

	schemaPath, err := createTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(schemaPath)
	}()

	subject := topic
	schemaRegistryContext, err := cmd.Flags().GetString("schema-registry-context")
	if err != nil {
		return err
	}
	if schemaRegistryContext != "" {
		subject = schemaRegistryContext
	}

	groupHandler := &GroupHandler{
		SrClient:    srClient,
		KeyFormat:   keyFormat,
		ValueFormat: valueFormat,
		Out:         cmd.OutOrStdout(),
		Subject:     subject,
		Properties: ConsumerProperties{
			Delimiter:   delimiter,
			FullHeader:  fullHeader,
			PrintKey:    printKey,
			PrintOffset: printOffset,
			SchemaPath:  schemaPath,
			Timestamp:   timestamp,
		},
	}
	return RunConsumer(consumer, groupHandler)
}

func (c *command) consumeOnPrem(cmd *cobra.Command, args []string) error {
	printKey, err := cmd.Flags().GetBool("print-key")
	if err != nil {
		return err
	}

	printOffset, err := cmd.Flags().GetBool("print-offset")
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

	keyFormat, err := cmd.Flags().GetString("key-format")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
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

	consumer, err := newOnPremConsumer(cmd, c.clientID, configFile, config)
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.FailedToCreateConsumerErrorMsg, err),
			errors.OnPremConfigGuideSuggestions,
		)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	if err := c.refreshOAuthBearerToken(cmd, consumer); err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	if err := ValidateTopic(adminClient, topicName); err != nil {
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
	if err := consumer.Subscribe(topicName, rebalanceCallback); err != nil {
		return err
	}

	output.ErrPrintln(c.Config.EnableColor, errors.StartingConsumerMsg)

	var srClient *schemaregistry.Client
	if slices.Contains(serdes.SchemaBasedFormats, valueFormat) {
		srClient, err = c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return err
		}
	}

	dir, err := createTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	groupHandler := &GroupHandler{
		SrClient:    srClient,
		KeyFormat:   keyFormat,
		ValueFormat: valueFormat,
		Out:         cmd.OutOrStdout(),
		Properties: ConsumerProperties{
			Delimiter:   delimiter,
			FullHeader:  fullHeader,
			PrintKey:    printKey,
			PrintOffset: printOffset,
			SchemaPath:  dir,
			Timestamp:   timestamp,
		},
	}
	return RunConsumer(consumer, groupHandler)
}
