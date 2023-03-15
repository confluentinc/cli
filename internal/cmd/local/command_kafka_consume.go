package local

import (
	"context"
	"fmt"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func (c *localCommand) newConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consume",
		Short: "---",
		Long:  "---",
		Args:  cobra.NoArgs,
		RunE:  c.consume,
	}

	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	return cmd
}

func (c *localCommand) consume(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	printKey, err := cmd.Flags().GetBool("print-key")
	if err != nil {
		return err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return err
	}

	fromBeginning, err := cmd.Flags().GetBool("from-beginning")
	if err != nil {
		return err
	}

	consumer, err := newOnPremConsumer(":"+c.Config.LocalPorts.PlaintextPort, fromBeginning)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateConsumerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	log.CliLogger.Tracef("Create consumer succeeded")

	if err := consumer.Subscribe(testTopicName, nil); err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingConsumerMsg)

	groupHandler := &kafka.GroupHandler{
		Ctx:    ctx,
		Out:    cmd.OutOrStdout(),
		Format: "string",
		Properties: kafka.ConsumerProperties{
			PrintKey:  printKey,
			Delimiter: delimiter,
		},
	}
	return kafka.RunConsumer(consumer, groupHandler)
}

func newOnPremConsumer(bootstrap string, fromBeginning bool) (*ckafka.Consumer, error) {
	group := fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"group.id":                              group,
		"client.id":                             "confluent-local",
		"bootstrap.servers":                     bootstrap,
		"partition.assignment.strategy":         "cooperative-sticky",
		"security.protocol":                     "PLAINTEXT",
		"auto.offset.reset":                     "latest",
	}
	log.CliLogger.Debugf("Created consumer group: %s", group)

	if fromBeginning {
		if err := configMap.SetKey("auto.offset.reset", "earliest"); err != nil {
			return nil, err
		}
	}

	switch log.CliLogger.Level {
	case log.DEBUG:
		if err := configMap.Set("debug=broker, topic, msg, protocol"); err != nil {
			return nil, err
		}
	case log.TRACE:
		if err := configMap.Set("debug=all"); err != nil {
			return nil, err
		}

	}
	return ckafka.NewConsumer(configMap)
}
