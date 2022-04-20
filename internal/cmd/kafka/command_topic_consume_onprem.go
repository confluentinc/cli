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
		RunE:  pcmd.NewCLIRunE(c.onPremConsume),
		Short: "Consume messages from a Kafka topic.",
		Long:  "Consume messages from a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.",
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
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema.")
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

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	configMap, err := getOnPremConsumerConfigMap(cmd, c.clientID)
	if err != nil {
		return err
	}
	consumerGroup, err := configMap.Get("group.id", "")
	if err != nil {
		return err
	}
	log.CliLogger.Debugf("Created consumer group: %s", consumerGroup)

	var srClient *srsdk.APIClient
	var ctx context.Context
	if valueFormat != "string" {
		// Only initialize client and context when schema is specified.
		if c.State == nil { // require log-in to use oauthbearer token
			return errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestion)
		}
		srClient, ctx, err = sr.GetSrApiClientWithToken(cmd, nil, c.Version, c.AuthToken())
		if err != nil {
			return err
		}
	}

	consumer, err := ckafka.NewConsumer(configMap)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateConsumerMsg, err).Error(), errors.OnPremConfigGuideSuggestion)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	err = c.refreshOAuthBearerToken(cmd, consumer)
	if err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	err = c.validateTopic(adminClient, topicName)
	if err != nil {
		return err
	}

	err = consumer.Subscribe(topicName, nil)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingConsumerMsg)

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
		Properties: ConsumerProperties{PrintKey: printKey, Delimiter: delimiter, SchemaPath: dir},
	}
	return runConsumer(cmd, consumer, groupHandler)
}
