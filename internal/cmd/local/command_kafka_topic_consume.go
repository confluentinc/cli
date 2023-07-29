package local

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *Command) newKafkaTopicConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consume <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kafkaTopicConsume,
		Short: "Consume messages from a Kafka topic.",
		Long:  "Consume messages from a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.\n\nTruncated message headers will be printed if they exist.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume message from topic "test" from the beginning and with keys printed.`,
				Code: "confluent local kafka topic consume test --from-beginning --print-key",
			},
		),
	}

	cmd.Flags().String("group", "", "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().Int64("offset", 0, "The offset from the beginning to consume from.")
	cmd.Flags().Int32("partition", -1, "The partition to consume from.")
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("timestamp", false, "Print message timestamp in milliseconds.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	pcmd.AddConsumerConfigFileFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("from-beginning", "offset")
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

	return cmd
}

func (c *Command) kafkaTopicConsume(cmd *cobra.Command, args []string) error {
	printKey, err := cmd.Flags().GetBool("print-key")
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

	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	consumer, err := newOnPremConsumer(cmd, ":"+c.Config.LocalPorts.PlaintextPort)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateConsumerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	err = kafka.ValidateTopic(adminClient, topicName)
	if err != nil {
		return err
	}

	offset, err := kafka.GetOffsetWithFallback(cmd)
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}
	partitionFilter := kafka.PartitionFilter{
		Changed: cmd.Flags().Changed("partition"),
		Index:   partition,
	}

	rebalanceCallback := kafka.GetRebalanceCallback(offset, partitionFilter)
	if err := consumer.Subscribe(topicName, rebalanceCallback); err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingConsumerMsg)

	groupHandler := &kafka.GroupHandler{
		Out:    cmd.OutOrStdout(),
		Format: "string",
		Properties: kafka.ConsumerProperties{
			PrintKey:  printKey,
			Timestamp: timestamp,
			Delimiter: delimiter,
		},
	}
	return kafka.RunConsumer(consumer, groupHandler)
}

func newOnPremConsumer(cmd *cobra.Command, bootstrap string) (*ckafka.Consumer, error) {
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}
	if group == "" {
		group = fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	}
	log.CliLogger.Debugf("Created consumer group: %s", group)

	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"group.id":                              group,
		"client.id":                             "confluent-local",
		"bootstrap.servers":                     bootstrap,
		"partition.assignment.strategy":         "cooperative-sticky",
		"security.protocol":                     "PLAINTEXT",
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return nil, err
	}

	if err := kafka.OverwriteKafkaClientConfigs(configMap, configFile, config); err != nil {
		return nil, err
	}

	if err := kafka.SetConsumerDebugOption(configMap); err != nil {
		return nil, err
	}

	return ckafka.NewConsumer(configMap)
}
