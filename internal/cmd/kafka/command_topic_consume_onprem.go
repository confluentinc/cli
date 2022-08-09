package kafka

import (
	"context"
	"fmt"
	"os"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *authenticatedTopicCommand) newConsumeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consume <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.onPremConsume,
		Short: "Consume messages from a Kafka topic.",
		Long:  "Consume messages from a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.\n\nTruncated message headers will be printed if they exist.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume message from topic "my_topic" with SSL protocol and SSL verification enabled (providing certificate and private key).`,
				Code: `confluent kafka topic consume my_topic --protocol SSL --bootstrap "localhost:19091" --ca-location my-cert.crt --cert-location client.pem --key-location client.key`},
			examples.Example{
				Text: `Consume message from topic "my_topic" with SASL_SSL/OAUTHBEARER protocol enabled (using MDS token).`,
				Code: `confluent kafka topic consume my_topic --protocol SASL_SSL --sasl-mechanism OAUTHBEARER --bootstrap "localhost:19091" --ca-location my-cert.crt`},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet())
	pcmd.AddProtocolFlag(cmd)
	pcmd.AddMechanismFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("group", "", "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().Int64("offset", 0, "The offset from the beginning to consume from.")
	cmd.Flags().Int32("partition", -1, "The partition to consume from.")
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("full-header", false, "Print complete content of message headers.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in json or avro format) for the consumer client.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("bootstrap")
	_ = cmd.MarkFlagRequired("ca-location")

	return cmd
}

func (c *authenticatedTopicCommand) onPremConsume(cmd *cobra.Command, args []string) error {
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

	valueFormat, err := cmd.Flags().GetString("value-format")
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

	consumer, err := newOnPremConsumer(cmd, c.clientID, configFile, config)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateConsumerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	err = c.refreshOAuthBearerToken(cmd, consumer)
	if err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	err = c.validateTopic(adminClient, topicName)
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
	err = consumer.Subscribe(topicName, rebalanceCallback)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingConsumerMsg)

	var srClient *srsdk.APIClient
	var ctx context.Context
	if valueFormat != "string" {
		// Only initialize client and context when schema is specified.
		if c.State == nil { // require log-in to use oauthbearer token
			return errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}
		srClient, ctx, err = sr.GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
		if err != nil {
			return err
		}
	}

	dir, err := sr.CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	groupHandler := &GroupHandler{
		SrClient:   srClient,
		Ctx:        ctx,
		Format:     valueFormat,
		Out:        cmd.OutOrStdout(),
		Properties: ConsumerProperties{PrintKey: printKey, FullHeader: fullHeader, Delimiter: delimiter, SchemaPath: dir},
	}
	return runConsumer(cmd, consumer, groupHandler)
}
