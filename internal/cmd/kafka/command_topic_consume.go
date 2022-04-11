package kafka

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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

func (c *hasAPIKeyTopicCommand) newConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "consume <topic>",
		Short:       "Consume messages from a Kafka topic.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.consume),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Consume items from the "my_topic" topic and press "Ctrl+C" to exit.`,
				Code: "confluent kafka topic consume -b my_topic",
			},
		),
	}
	cmd.Flags().String("group", fmt.Sprintf("confluent_cli_consumer_%s", uuid.New()), "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema. Note that schema references are not supported for avro.")
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().Bool("full-header", false, "Print complete content of message headers.")
	cmd.Flags().String("delimiter", "\t", "The delimiter separating each key and value.")
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
	beginning, err := cmd.Flags().GetBool("from-beginning")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
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

	consumer, err := NewConsumer(group, cluster, c.clientID, beginning)
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

	utils.ErrPrintln(cmd, errors.StartingConsumerMsg)

	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	err = consumer.Subscribe(topic, nil)
	if err != nil {
		return err
	}

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
