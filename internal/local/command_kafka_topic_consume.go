package local

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/internal/kafka"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newKafkaTopicConsumeCommand() *cobra.Command {
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
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the consumer client. For a full list, see https://docs.confluent.io/platform/current/clients/librdkafka/html/md_CONFIGURATION.html`)
	pcmd.AddConsumerConfigFileFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("from-beginning", "offset")
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

	return cmd
}

func (c *command) kafkaTopicConsume(cmd *cobra.Command, args []string) error {
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
	consumer, err := newOnPremConsumer(cmd, c.getPlaintextBootstrapServers())
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.FailedToCreateConsumerErrorMsg, err),
			errors.OnPremConfigGuideSuggestions,
		)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	adminClient, err := ckgo.NewAdminClientFromConsumer(consumer)
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
	if cmd.Flags().Changed("group") && !cmd.Flags().Changed("from-beginning") && !cmd.Flags().Changed("offset") {
		rebalanceCallback = nil
	}
	if err := consumer.Subscribe(topicName, rebalanceCallback); err != nil {
		return err
	}

	output.ErrPrintln(c.Config.EnableColor, errors.StartingConsumerMsg)

	groupHandler := &kafka.GroupHandler{
		Out:         cmd.OutOrStdout(),
		KeyFormat:   "string",
		ValueFormat: "string",
		Properties: kafka.ConsumerProperties{
			PrintKey:  printKey,
			Timestamp: timestamp,
			Delimiter: delimiter,
		},
	}
	return runConsumer(consumer, groupHandler)
}

func newOnPremConsumer(cmd *cobra.Command, bootstrap string) (*ckgo.Consumer, error) {
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}
	if group == "" {
		group = fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	}
	log.CliLogger.Debugf("Created consumer group: %s", group)

	configMap := &ckgo.ConfigMap{
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

	return ckgo.NewConsumer(configMap)
}

func runConsumer(consumer *ckgo.Consumer, groupHandler *kafka.GroupHandler) error {
	run := true
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for run {
		select {
		case <-signals: // Trap SIGINT to trigger a shutdown.
			output.ErrPrintln(false, "Stopping Consumer.")
			if _, err := consumer.Commit(); err != nil {
				log.CliLogger.Warnf("Failed to commit current consumer offset: %v", err)
			}
			consumer.Close()
			run = false
		default:
			event := consumer.Poll(100) // polling event from consumer with a timeout of 100ms
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *ckgo.Message:
				if err := kafka.ConsumeMessage(e, groupHandler); err != nil {
					commitErrCh := make(chan error, 1)
					go func() {
						_, err := consumer.Commit()
						commitErrCh <- err
					}()
					select {
					case commitErr := <-commitErrCh:
						if commitErr != nil {
							log.CliLogger.Warnf("Failed to commit current consumer offset: %v", commitErr)
						}
					// Time out in case consumer has lost connection to Kafka and commit would hang
					case <-time.After(5 * time.Second):
						log.CliLogger.Warnf("Commit operation timed out")
					}

					return err
				}
			case ckgo.Error:
				fmt.Fprintf(groupHandler.Out, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == ckgo.ErrAllBrokersDown {
					run = false
				}
			}
		}
	}
	return nil
}
