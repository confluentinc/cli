package kafka

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/serdes"
)

func (c *command) newProduceCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "produce <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.produceOnPrem,
		Short: "Produce messages to a Kafka topic.",
		Long:  "Produce messages to a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.\n\nWhen using this command, you cannot modify the message header, and the message header will not be printed out.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Produce message to topic "my_topic" with SASL_SSL/PLAIN protocol (providing username and password).`,
				Code: `confluent kafka topic produce my_topic --protocol SASL_SSL --sasl-mechanism PLAIN --bootstrap "localhost:19091" --username user --password secret --ca-location my-cert.crt`,
			},
			examples.Example{
				Text: `Produce message to topic "my_topic" with SSL protocol, and SSL verification enabled.`,
				Code: `confluent kafka topic produce my_topic --protocol SSL --bootstrap "localhost:18091" --ca-location my-cert.crt`,
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet())
	pcmd.AddProtocolFlag(cmd)
	pcmd.AddMechanismFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("key-schema", "", "The filepath of the message key schema.")
	cmd.Flags().String("schema", "", "The filepath of the message value schema.")
	pcmd.AddKeyFormatFlag(cmd)
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	pcmd.AddProducerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-endpoint", "", "The URL of the Schema Registry cluster.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cobra.CheckErr(cmd.MarkFlagRequired("bootstrap"))
	cobra.CheckErr(cmd.MarkFlagRequired("ca-location"))

	cmd.MarkFlagsMutuallyExclusive("config-file", "config")

	return cmd
}

func (c *command) produceOnPrem(cmd *cobra.Command, args []string) error {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	producer, err := newOnPremProducer(cmd, c.clientID, configFile, config)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateProducerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	if err := c.refreshOAuthBearerToken(cmd, producer); err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topic := args[0]
	if err := ValidateTopic(adminClient, topic); err != nil {
		return err
	}

	keyFormat, keySubject, keySerializer, err := prepareSerializer(cmd, topic, "key")
	if err != nil {
		return err
	}

	valueFormat, valueSubject, valueSerializer, err := prepareSerializer(cmd, topic, "value")
	if err != nil {
		return err
	}

	keySchema, err := cmd.Flags().GetString("key-schema")
	if err != nil {
		return err
	}

	schema, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	refs, err := sr.ReadSchemaRefs(cmd)
	if err != nil {
		return err
	}

	dir, err := sr.CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	keySchemaConfigs := &sr.RegisterSchemaConfigs{
		Subject:    keySubject,
		SchemaDir:  dir,
		SchemaType: keySerializer.GetSchemaName(),
		Format:     keyFormat,
		SchemaPath: keySchema,
		Refs:       refs,
	}
	keyMetadata, keyReferencePathMap, err := c.registerSchemaOnPrem(cmd, keySchemaConfigs)
	if err != nil {
		return err
	}
	if err := keySerializer.LoadSchema(keySchema, keyReferencePathMap); err != nil {
		return err
	}

	valueSchemaConfigs := &sr.RegisterSchemaConfigs{
		Subject:    valueSubject,
		SchemaDir:  dir,
		SchemaType: valueSerializer.GetSchemaName(),
		Format:     valueFormat,
		SchemaPath: schema,
		Refs:       refs,
	}
	valueMetadata, referencePathMap, err := c.registerSchemaOnPrem(cmd, valueSchemaConfigs)
	if err != nil {
		return err
	}
	if err := valueSerializer.LoadSchema(schema, referencePathMap); err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingProducerMsg)

	var scanErr error
	input, scan := PrepareInputChannel(&scanErr)

	signals := make(chan os.Signal, 1) // Trap SIGINT to trigger a shutdown.
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(input)
	}()
	go scan() // Prime reader

	deliveryChan := make(chan ckafka.Event)
	for data := range input {
		if len(data) == 0 {
			go scan()
			continue
		}

		msg, err := GetProduceMessage(cmd, keyMetadata, valueMetadata, topic, data, keySerializer, valueSerializer)
		if err != nil {
			return err
		}
		if err := producer.Produce(msg, deliveryChan); err != nil {
			output.ErrPrintf(errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckafka.Message)           // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topic)
			if isProduceToCompactedTopicError {
				scanErr = err
				close(input)
				break
			}
			output.ErrPrintf(errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func prepareSerializer(cmd *cobra.Command, topic, mode string) (string, string, serdes.SerializationProvider, error) {
	valueFormat, err := cmd.Flags().GetString(fmt.Sprintf("%s-format", mode))
	if err != nil {
		return "", "", nil, err
	}

	serializer, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return "", "", nil, err
	}

	return valueFormat, topicNameStrategy(topic), serializer, nil
}

func (c *command) registerSchemaOnPrem(cmd *cobra.Command, schemaCfg *sr.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	metaInfo := []byte{}
	referencePathMap := map[string]string{}
	if schemaCfg.Format != "string" && len(schemaCfg.SchemaPath) > 0 {
		if c.State == nil { // require log-in to use oauthbearer token
			return nil, nil, errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}
		srClient, ctx, err := sr.GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
		if err != nil {
			return nil, nil, err
		}

		id, err := sr.RegisterSchemaWithAuth(cmd, schemaCfg, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
		metaInfo = sr.GetMetaInfoFromSchemaId(id)

		referencePathMap, err = sr.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	return metaInfo, referencePathMap, nil
}
