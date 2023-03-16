package local

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *kafkaCommand) newConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consume",
		Short: "---",
		Long:  "---",
		Args:  cobra.NoArgs,
		RunE:  c.consume,
	}

	cmd.Flags().String("group", "", "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().Int64("offset", 0, "The offset from the beginning to consume from.")
	cmd.Flags().Int32("partition", -1, "The partition to consume from.")
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("timestamp", false, "Print message timestamp in milliseconds.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in json or avro format) for the consumer client.")
	return cmd
}

func (c *kafkaCommand) consume(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

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

	consumer, err := newOnPremConsumer(cmd, ":"+c.Config.LocalPorts.PlaintextPort)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateConsumerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	if cmd.Flags().Changed("from-beginning") && cmd.Flags().Changed("offset") {
		return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "from-beginning", "offset")
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
	if err := consumer.Subscribe(testTopicName, rebalanceCallback); err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingConsumerMsg)

	groupHandler := &kafka.GroupHandler{
		Ctx:    ctx,
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

	if cmd.Flags().Changed("config-file") && cmd.Flags().Changed("config") {
		return nil, errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "config-file", "config")
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}
	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return nil, err
	}

	err = kafka.OverwriteKafkaClientConfigs(configMap, configFile, config)
	if err != nil {
		return nil, err
	}

	if err := kafka.SetConsumerDebugOption(configMap); err != nil {
		return nil, err
	}

	return ckafka.NewConsumer(configMap)
}
