package kafka

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *authenticatedTopicCommand) newProduceCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "produce <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.onPremProduce,
		Short: "Produce messages to a Kafka topic.",
		Long:  "Produce messages to a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.\n\nWhen using this command, you cannot modify the message header, and the message header will not be printed out.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Produce message to topic "my-topic" providing embedded Kafka REST Proxy endpoint.`,
				Code: "confluent kafka topic produce my-topic --url http://localhost:8090/kafka",
			},
			examples.Example{
				Text: `Produce messages to topic "my-topic" providing Kafka REST Proxy endpoint.`,
				Code: "confluent kafka topic produce my-topic --url http://localhost:8082",
			},
			examples.Example{
				Text: `Produce message to topic "my-topic" with SASL_SSL/PLAIN protocol (providing username and password).`,
				Code: `confluent kafka topic produce my-topic --protocol SASL_SSL --sasl-mechanism PLAIN --bootstrap "localhost:19091" --username user --password secret --ca-location my-cert.crt`,
			},
			examples.Example{
				Text: `Produce messages to topic "my-topic" with SSL protocol, and SSL verification enabled.`,
				Code: `confluent kafka topic produce my-topic --protocol SSL --bootstrap "localhost:18091" --ca-location my-cert.crt`,
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet())
	pcmd.AddProtocolFlag(cmd)
	pcmd.AddMechanismFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("schema", "", "The path to the local schema file.")
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in json or avro format) for the producer client.")
	cmd.Flags().String("schema-registry-endpoint", "", "The URL of the Schema Registry cluster.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("bootstrap")
	_ = cmd.MarkFlagRequired("ca-location")

	return cmd
}

func (c *authenticatedTopicCommand) onPremProduce(cmd *cobra.Command, args []string) error {
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

	producer, err := newOnPremProducer(cmd, c.clientID, configFile, config)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateProducerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	err = c.refreshOAuthBearerToken(cmd, producer)
	if err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	err = c.validateTopic(adminClient, topicName)
	if err != nil {
		return err
	}

	valueFormat, subject, serializationProvider, err := prepareSerializer(cmd, topicName)
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

	// Meta info contains magic byte and schema ID (4 bytes).
	schemaCfg := &sr.RegisterSchemaConfigs{
		Subject:     subject,
		SchemaDir:   dir,
		SchemaType:  serializationProvider.GetSchemaName(),
		ValueFormat: valueFormat,
		SchemaPath:  &schema,
		Refs:        refs,
	}
	metaInfo, referencePathMap, err := c.registerSchema(cmd, schemaCfg)
	if err != nil {
		return err
	}
	err = serializationProvider.LoadSchema(schema, referencePathMap)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingProducerMsg)

	// Line reader for producer input.
	scanner := bufio.NewScanner(os.Stdin)
	// On-prem Kafka messageMaxBytes: using the same value of cloud. TODO: allow larger sizes if customers request
	// https://github.com/confluentinc/cc-spec-kafka/blob/9f0af828d20e9339aeab6991f32d8355eb3f0776/plugins/kafka/kafka.go#L43.
	const maxScanTokenSize = 1024*1024*2 + 12
	scanner.Buffer(nil, maxScanTokenSize)
	input := make(chan string, 1)
	// Avoid blocking in for loop so ^C or ^D can exit immediately.
	var scanErr error
	scan := func() {
		hasNext := scanner.Scan()
		if !hasNext {
			// Actual error.
			if scanner.Err() != nil {
				scanErr = scanner.Err()
			}
			// Otherwise just EOF.
			close(input)
		} else {
			input <- scanner.Text()
		}
	}

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

		msg, err := getProduceMessage(cmd, metaInfo, topicName, data, serializationProvider)
		if err != nil {
			return err
		}
		err = producer.Produce(msg, deliveryChan)
		if err != nil {
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckafka.Message)           // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topicName)
			if isProduceToCompactedTopicError {
				scanErr = err
				close(input)
				break
			}
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func prepareSerializer(cmd *cobra.Command, topicName string) (string, string, serdes.SerializationProvider, error) {
	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return "", "", nil, err
	}
	subject := topicNameStrategy(topicName)
	serializationProvider, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return "", "", nil, err
	}
	return valueFormat, subject, serializationProvider, nil
}

func (c *authenticatedTopicCommand) registerSchema(cmd *cobra.Command, schemaCfg *sr.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	metaInfo := []byte{}
	referencePathMap := map[string]string{}
	if schemaCfg.ValueFormat != "string" && len(*schemaCfg.SchemaPath) > 0 {
		if c.State == nil { // require log-in to use oauthbearer token
			return nil, nil, errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}
		srClient, ctx, err := sr.GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
		if err != nil {
			return nil, nil, err
		}

		metaInfo, err = sr.RegisterSchemaWithAuth(cmd, schemaCfg, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
		referencePathMap, err = sr.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, srClient, ctx)
		if err != nil {
			return metaInfo, nil, err
		}
	}
	return metaInfo, referencePathMap, nil
}
