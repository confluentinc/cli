package kafka

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"time"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	schemaregistry "github.com/confluentinc/cli/internal/pkg/schema-registry"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	messageOffset = 5 // Schema ID is stored at the [1:5] bytes of a message as meta info (when valid)

	// required fields of SASL/oauthbearer configuration
	principalClaimNameKey = "principalClaimName"
	principalKey          = "principal"
	oauthConfig           = "principalClaimName=confluent principal=admin"
)

var (
	// Regex for sasl.oauthbearer.config, which constrains it to be
	// 1 or more name=value pairs with optional ignored whitespace
	oauthbearerConfigRegex = regexp.MustCompile(`^(\s*(\w+)\s*=\s*(\w+))+\s*$`)
	// Regex used to extract name=value pairs from sasl.oauthbearer.config
	oauthbearerNameEqualsValueRegex = regexp.MustCompile(`(\w+)\s*=\s*(\w+)`)
)

type ConsumerProperties struct {
	Delimiter  string
	FullHeader bool
	PrintKey   bool
	Timestamp  bool
	SchemaPath string
}

// GroupHandler instances are used to handle individual topic-partition claims.
type GroupHandler struct {
	SrClient   *schemaregistry.Client
	Format     string
	Out        io.Writer
	Subject    string
	Properties ConsumerProperties
}

func (c *command) refreshOAuthBearerToken(cmd *cobra.Command, client ckafka.Handle) error {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return err
	}
	saslMechanism, err := cmd.Flags().GetString("sasl-mechanism")
	if err != nil {
		return err
	}
	if protocol == "SASL_SSL" && saslMechanism == "OAUTHBEARER" {
		oart := ckafka.OAuthBearerTokenRefresh{Config: oauthConfig}
		if c.State == nil { // require log-in to use oauthbearer token
			return errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}
		oauthBearerToken, retrieveErr := retrieveUnsecuredToken(oart, c.AuthToken())
		if retrieveErr != nil {
			_ = client.SetOAuthBearerTokenFailure(retrieveErr.Error())
			return fmt.Errorf("token retrieval error: %w", retrieveErr)
		} else {
			setTokenError := client.SetOAuthBearerToken(oauthBearerToken)
			if setTokenError != nil {
				_ = client.SetOAuthBearerTokenFailure(setTokenError.Error())
				return fmt.Errorf("error setting token and extensions: %w", setTokenError)
			}
		}
	}
	return nil
}

func retrieveUnsecuredToken(e ckafka.OAuthBearerTokenRefresh, tokenValue string) (ckafka.OAuthBearerToken, error) {
	config := e.Config
	if !oauthbearerConfigRegex.MatchString(config) {
		return ckafka.OAuthBearerToken{}, fmt.Errorf("ignoring event %T due to malformed config: %s", e, config)
	}
	oauthbearerConfigMap := map[string]string{
		principalClaimNameKey: "sub",
	}
	for _, kv := range oauthbearerNameEqualsValueRegex.FindAllStringSubmatch(config, -1) {
		oauthbearerConfigMap[kv[1]] = kv[2]
	}
	principal := oauthbearerConfigMap[principalKey]
	if principal == "" {
		return ckafka.OAuthBearerToken{}, fmt.Errorf("ignoring event %T: no %s: %s", e, principalKey, config)
	}

	if len(oauthbearerConfigMap) > 2 { // do not proceed if there are any unknown name=value pairs
		return ckafka.OAuthBearerToken{}, fmt.Errorf("ignoring event %T: unrecognized key(s): %s", e, config)
	}

	now := time.Now()
	expiration := now.Add(time.Second * time.Duration(3600)) // timeout after 60 mins. TODO: re-authenticate after timout
	oauthBearerToken := ckafka.OAuthBearerToken{
		TokenValue: tokenValue,
		Expiration: expiration,
		Principal:  principal,
	}
	return oauthBearerToken, nil
}

func newProducer(kafka *configv1.KafkaClusterConfig, clientID, configPath string, configStrings []string) (*ckafka.Producer, error) {
	configMap, err := getProducerConfigMap(kafka, clientID)
	if err != nil {
		return nil, err
	}

	return newProducerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newConsumer(group string, kafka *configv1.KafkaClusterConfig, clientID, configPath string, configStrings []string) (*ckafka.Consumer, error) {
	configMap, err := getConsumerConfigMap(group, kafka, clientID)
	if err != nil {
		return nil, err
	}

	return newConsumerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newOnPremProducer(cmd *cobra.Command, clientID, configPath string, configStrings []string) (*ckafka.Producer, error) {
	configMap, err := getOnPremProducerConfigMap(cmd, clientID)
	if err != nil {
		return nil, err
	}

	return newProducerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newOnPremConsumer(cmd *cobra.Command, clientID, configPath string, configStrings []string) (*ckafka.Consumer, error) {
	configMap, err := getOnPremConsumerConfigMap(cmd, clientID)
	if err != nil {
		return nil, err
	}

	return newConsumerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

// example: https://github.com/confluentinc/confluent-kafka-go/blob/e01dd295220b5bf55f3fbfabdf8cc6d3f0ae185f/examples/cooperative_consumer_example/cooperative_consumer_example.go#L121
func GetRebalanceCallback(offset ckafka.Offset, partitionFilter PartitionFilter) func(*ckafka.Consumer, ckafka.Event) error {
	return func(consumer *ckafka.Consumer, event ckafka.Event) error {
		switch ev := event.(type) { // ev is of type ckafka.Event
		case ckafka.AssignedPartitions:
			partitions := make([]ckafka.TopicPartition, len(ev.Partitions))
			for i, partition := range ev.Partitions {
				partition.Offset = offset
				partitions[i] = partition
			}
			partitions = getPartitionsByIndex(partitions, partitionFilter)

			if err := consumer.IncrementalAssign(partitions); err != nil {
				return err
			}
		case ckafka.RevokedPartitions:
			if consumer.AssignmentLost() {
				output.ErrPrintln("%% Current assignment lost.")
			}
			parts := getPartitionsByIndex(ev.Partitions, partitionFilter)
			if err := consumer.IncrementalUnassign(parts); err != nil {
				return err
			}
		}
		return nil
	}
}

func consumeMessage(e *ckafka.Message, h *GroupHandler) error {
	value := e.Value
	if h.Properties.PrintKey {
		key := e.Key
		var keyString string
		if len(key) == 0 {
			keyString = "null"
		} else {
			keyString = string(key)
		}
		if _, err := fmt.Fprint(h.Out, keyString+h.Properties.Delimiter); err != nil {
			return err
		}
	}

	deserializationProvider, err := serdes.GetDeserializationProvider(h.Format)
	if err != nil {
		return err
	}

	if h.Format != "string" {
		schemaPath, referencePathMap, err := h.RequestSchema(value)
		if err != nil {
			return err
		}
		// Message body is encoded after 5 bytes of meta information.
		value = value[messageOffset:]
		if err := deserializationProvider.LoadSchema(schemaPath, referencePathMap); err != nil {
			return err
		}
	}
	jsonMessage, err := serdes.Deserialize(deserializationProvider, value)
	if err != nil {
		return err
	}

	if h.Properties.Timestamp {
		jsonMessage = fmt.Sprintf("Timestamp: %d\t%s", e.Timestamp.UnixMilli(), jsonMessage)
	}

	if _, err := fmt.Fprintln(h.Out, jsonMessage); err != nil {
		return err
	}

	if e.Headers != nil {
		var headers any = e.Headers
		if h.Properties.FullHeader {
			headers = getFullHeaders(e.Headers)
		}
		if _, err := fmt.Fprintf(h.Out, "%% Headers: %v\n", headers); err != nil {
			return err
		}
	}

	return nil
}

func RunConsumer(consumer *ckafka.Consumer, groupHandler *GroupHandler) error {
	run := true
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for run {
		select {
		case <-signals: // Trap SIGINT to trigger a shutdown.
			output.ErrPrintln(errors.StoppingConsumerMsg)
			consumer.Close()
			run = false
		default:
			event := consumer.Poll(100) // polling event from consumer with a timeout of 100ms
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *ckafka.Message:
				if err := consumeMessage(e, groupHandler); err != nil {
					return err
				}
			case ckafka.Error:
				fmt.Fprintf(groupHandler.Out, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == ckafka.ErrAllBrokersDown {
					run = false
				}
			}
		}
	}
	return nil
}

func (h *GroupHandler) RequestSchema(value []byte) (string, map[string]string, error) {
	if len(value) == 0 || value[0] != 0x0 {
		return "", nil, errors.NewErrorWithSuggestions("unknown magic byte", fmt.Sprintf("Check that all messages from this topic are in the %s format.", h.Format))
	}
	if len(value) < messageOffset {
		return "", nil, errors.New("failed to find schema ID in topic data")
	}

	// Retrieve schema from cluster only if schema is specified.
	schemaID := int32(binary.BigEndian.Uint32(value[1:messageOffset])) // schema id is stored as a part of message meta info

	// Create temporary file to store schema retrieved (also for cache). Retry if get error retrieving schema or writing temp schema file
	tempStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%s-%d.txt", h.Subject, schemaID))
	tempRefStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%s-%d.ref", h.Subject, schemaID))
	var references []srsdk.SchemaReference
	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		opts := &srsdk.GetSchemaOpts{Subject: optional.NewString(h.Subject)}
		schemaString, err := h.SrClient.GetSchema(schemaID, opts)
		if err != nil {
			return "", nil, err
		}
		if err := os.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644); err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		if err := os.WriteFile(tempRefStorePath, refBytes, 0644); err != nil {
			return "", nil, err
		}
		references = schemaString.References
	} else {
		refBlob, err := os.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		if err := json.Unmarshal(refBlob, &references); err != nil {
			return "", nil, err
		}
	}

	// Store the references in temporary files
	referencePathMap, err := sr.StoreSchemaReferences(h.Properties.SchemaPath, references, h.SrClient)
	if err != nil {
		return "", nil, err
	}

	return tempStorePath, referencePathMap, nil
}

func getFullHeaders(headers []ckafka.Header) []string {
	headerStrings := make([]string, len(headers))
	for i, header := range headers {
		headerStrings[i] = getHeaderString(header)
	}
	return headerStrings
}

func getHeaderString(header ckafka.Header) string {
	if header.Value == nil {
		return fmt.Sprintf("%s=nil", header.Key)
	} else if len(header.Value) == 0 {
		return fmt.Sprintf("%s=<empty>", header.Key)
	} else {
		return fmt.Sprintf(`%s="%s"`, header.Key, string(header.Value))
	}
}
