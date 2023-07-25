package kafka

import (
	"context"
	"fmt"
	"os"

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

func (c *command) newConsumeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consume <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.consumeOnPrem,
		Short: "Consume messages from a Kafka topic.",
		Long:  "Consume messages from a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.\n\nTruncated message headers will be printed if they exist.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume message from topic "my_topic" with SSL protocol and SSL verification enabled (providing certificate and private key).`,
				Code: `confluent kafka topic consume my_topic --protocol SSL --bootstrap "localhost:19091" --ca-location my-cert.crt --cert-location client.pem --key-location client.key`,
			},
			examples.Example{
				Text: `Consume message from topic "my_topic" with SASL_SSL/OAUTHBEARER protocol enabled (using MDS token).`,
				Code: `confluent kafka topic consume my_topic --protocol SASL_SSL --sasl-mechanism OAUTHBEARER --bootstrap "localhost:19091" --ca-location my-cert.crt`,
			},
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
	cmd.Flags().Bool("timestamp", false, "Print message timestamp in milliseconds.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	pcmd.AddConsumerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-endpoint", "", "The URL of the Schema Registry cluster.")

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))
	cobra.CheckErr(cmd.MarkFlagRequired("bootstrap"))
	cobra.CheckErr(cmd.MarkFlagRequired("ca-location"))

	cmd.MarkFlagsMutuallyExclusive("config", "config-file")
	cmd.MarkFlagsMutuallyExclusive("from-beginning", "offset")

	return cmd
}

func (c *command) consumeOnPrem(cmd *cobra.Command, args []string) error {
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
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateConsumerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
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

	output.ErrPrintln(errors.StartingConsumerMsg)

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
		SrClient: srClient,
		Ctx:      ctx,
		Format:   valueFormat,
		Out:      cmd.OutOrStdout(), // TODO
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
