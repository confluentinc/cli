package kafka

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
	"github.com/confluentinc/cli/v4/pkg/serdes"
)

const (
	// required fields of SASL/oauthbearer configuration
	principalClaimNameKey = "principalClaimName"
	principalKey          = "principal"
	oauthConfig           = "principalClaimName=confluent principal=admin"
	keySchemaHeaderKey    = "__key_schema_id"
	valueSchemaHeaderKey  = "__value_schema_id"
)

var (
	// Regex for sasl.oauthbearer.config, which constrains it to be
	// 1 or more name=value pairs with optional ignored whitespace
	oauthbearerConfigRegex = regexp.MustCompile(`^(\s*(\w+)\s*=\s*(\w+))+\s*$`)
	// Regex used to extract name=value pairs from sasl.oauthbearer.config
	oauthbearerNameEqualsValueRegex = regexp.MustCompile(`(\w+)\s*=\s*(\w+)`)
)

type ConsumerProperties struct {
	Delimiter   string
	FullHeader  bool
	PrintKey    bool
	PrintOffset bool
	Timestamp   bool
	SchemaPath  string
}

// GroupHandler instances are used to handle individual topic-partition claims.
type GroupHandler struct {
	SrClient                 *schemaregistry.Client
	SrApiKey                 string
	SrApiSecret              string
	SrClusterId              string
	SrClusterEndpoint        string
	Token                    string
	CertificateAuthorityPath string
	ClientCertPath           string
	ClientKeyPath            string
	KeyFormat                string
	ValueFormat              string
	Out                      io.Writer
	Subject                  string
	Topic                    string
	Properties               ConsumerProperties
}

func (c *command) refreshOAuthBearerToken(cmd *cobra.Command, client ckgo.Handle, oart ckgo.OAuthBearerTokenRefresh) error {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return err
	}
	saslMechanism, err := cmd.Flags().GetString("sasl-mechanism")
	if err != nil {
		return err
	}
	if protocol == "SASL_SSL" && saslMechanism == "OAUTHBEARER" {
		if c.Context.GetState() == nil { // require log-in to use oauthbearer token
			return errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}
		err := c.mdsRequestAndAuthTokenUpdate(cmd)
		if err != nil {
			return err
		}
		oauthBearerToken, retrieveErr := c.retrieveUnsecuredToken(oart)
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

func (c *command) mdsRequestAndAuthTokenUpdate(cmd *cobra.Command) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	req := mdsv1.ExtendAuthRequest{
		AccessToken:  c.Context.GetAuthToken(),
		RefreshToken: c.Context.GetAuthRefreshToken(),
	}
	resp, _, err := client.SSODeviceAuthorizationApi.ExtendDeviceAuth(context.Background(), req)
	if err != nil {
		return err
	}
	c.Context.State.AuthToken = resp.AuthToken
	err = c.Context.Save()
	if err != nil {
		return err
	}
	return nil
}

func (c *command) retrieveUnsecuredToken(e ckgo.OAuthBearerTokenRefresh) (ckgo.OAuthBearerToken, error) {
	config := e.Config
	if !oauthbearerConfigRegex.MatchString(config) {
		return ckgo.OAuthBearerToken{}, fmt.Errorf("ignoring event %T due to malformed config: %s", e, config)
	}
	oauthbearerConfigMap := map[string]string{
		principalClaimNameKey: "sub",
	}
	for _, kv := range oauthbearerNameEqualsValueRegex.FindAllStringSubmatch(config, -1) {
		oauthbearerConfigMap[kv[1]] = kv[2]
	}
	principal := oauthbearerConfigMap[principalKey]
	if principal == "" {
		return ckgo.OAuthBearerToken{}, fmt.Errorf("ignoring event %T: no %s: %s", e, principalKey, config)
	}

	if len(oauthbearerConfigMap) > 2 { // do not proceed if there are any unknown name=value pairs
		return ckgo.OAuthBearerToken{}, fmt.Errorf("ignoring event %T: unrecognized key(s): %s", e, config)
	}

	expClaim, err := jwt.GetClaim(c.Context.GetAuthToken(), "exp")
	if err != nil {
		return ckgo.OAuthBearerToken{}, err
	}
	exp, ok := expClaim.(float64)
	if !ok {
		return ckgo.OAuthBearerToken{}, fmt.Errorf(errors.MalformedTokenErrorMsg, "exp")
	}
	expiration := time.Unix(int64(exp), 0)
	oauthBearerToken := ckgo.OAuthBearerToken{
		TokenValue: c.Context.GetAuthToken(),
		Expiration: expiration,
		Principal:  principal,
	}
	return oauthBearerToken, nil
}

func newProducer(kafka *config.KafkaClusterConfig, clientID, configPath string, configStrings []string) (*ckgo.Producer, error) {
	configMap, err := getProducerConfigMap(kafka, clientID)
	if err != nil {
		return nil, fmt.Errorf(errors.FailedToGetConfigurationErrorMsg, err)
	}

	return newProducerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newConsumer(group string, kafka *config.KafkaClusterConfig, clientID, configPath string, configStrings []string) (*ckgo.Consumer, error) {
	configMap, err := getConsumerConfigMap(group, kafka, clientID)
	if err != nil {
		return nil, fmt.Errorf(errors.FailedToGetConfigurationErrorMsg, err)
	}

	return newConsumerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newOnPremProducer(cmd *cobra.Command, clientID, configPath string, configStrings []string) (*ckgo.Producer, error) {
	configMap, err := getOnPremProducerConfigMap(cmd, clientID)
	if err != nil {
		return nil, fmt.Errorf(errors.FailedToGetConfigurationErrorMsg, err)
	}

	return newProducerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newOnPremConsumer(cmd *cobra.Command, clientID, configPath string, configStrings []string) (*ckgo.Consumer, error) {
	configMap, err := getOnPremConsumerConfigMap(cmd, clientID)
	if err != nil {
		return nil, fmt.Errorf(errors.FailedToGetConfigurationErrorMsg, err)
	}

	return newConsumerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

// example: https://github.com/confluentinc/confluent-kafka-go/blob/e01dd295220b5bf55f3fbfabdf8cc6d3f0ae185f/examples/cooperative_consumer_example/cooperative_consumer_example.go#L121
func GetRebalanceCallback(offset ckgo.Offset, partitionFilter PartitionFilter) func(*ckgo.Consumer, ckgo.Event) error {
	return func(consumer *ckgo.Consumer, event ckgo.Event) error {
		switch ev := event.(type) { // ev is of type ckafka.Event
		case ckgo.AssignedPartitions:
			partitions := make([]ckgo.TopicPartition, len(ev.Partitions))
			for i, partition := range ev.Partitions {
				partition.Offset = offset
				partitions[i] = partition
			}
			partitions = getPartitionsByIndex(partitions, partitionFilter)

			if err := consumer.IncrementalAssign(partitions); err != nil {
				return err
			}
		case ckgo.RevokedPartitions:
			if consumer.AssignmentLost() {
				output.ErrPrintln(false, "%% Current assignment lost.")
			}
			parts := getPartitionsByIndex(ev.Partitions, partitionFilter)
			if err := consumer.IncrementalUnassign(parts); err != nil {
				return err
			}
		}
		return nil
	}
}

func ConsumeMessage(message *ckgo.Message, h *GroupHandler) error {
	srAuth := serdes.SchemaRegistryAuth{
		ApiKey:                   h.SrApiKey,
		ApiSecret:                h.SrApiSecret,
		CertificateAuthorityPath: h.CertificateAuthorityPath,
		ClientCertPath:           h.ClientCertPath,
		ClientKeyPath:            h.ClientKeyPath,
		Token:                    h.Token,
	}

	if h.Properties.PrintKey {
		keyDeserializer, err := serdes.GetDeserializationProvider(h.KeyFormat)
		if err != nil {
			return err
		}

		err = keyDeserializer.InitDeserializer(h.SrClusterEndpoint, h.SrClusterId, "key", srAuth, nil)
		if err != nil {
			return err
		}

		if err := keyDeserializer.LoadSchema(h.Subject, h.Properties.SchemaPath, serde.KeySerde, message); err != nil {
			return err
		}

		jsonMessage, err := keyDeserializer.Deserialize(h.Topic, message.Headers, message.Key)
		if err != nil {
			return err
		}
		if jsonMessage == "" {
			jsonMessage = "null"
		}

		if _, err := fmt.Fprint(h.Out, jsonMessage+h.Properties.Delimiter); err != nil {
			return err
		}
	}

	valueDeserializer, err := serdes.GetDeserializationProvider(h.ValueFormat)
	if err != nil {
		return err
	}

	err = valueDeserializer.InitDeserializer(h.SrClusterEndpoint, h.SrClusterId, "value", srAuth, nil)
	if err != nil {
		return err
	}

	if err := valueDeserializer.LoadSchema(h.Subject, h.Properties.SchemaPath, serde.ValueSerde, message); err != nil {
		return err
	}

	messageString, err := getMessageString(message, valueDeserializer, h.Properties, h.Topic)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(h.Out, messageString); err != nil {
		return err
	}

	if message.Headers != nil {
		message.Headers = unmarshalSchemaIdHeader(message.Headers)
		var headers any = message.Headers
		if h.Properties.FullHeader {
			headers = getFullHeaders(message.Headers)
		}
		if _, err := fmt.Fprintf(h.Out, "%% Headers: %v\n", headers); err != nil {
			return err
		}
	}

	return nil
}

func getMessageString(message *ckgo.Message, valueDeserializer serdes.DeserializationProvider, properties ConsumerProperties, topic string) (string, error) {
	messageString, err := valueDeserializer.Deserialize(topic, message.Headers, message.Value)
	if err != nil {
		return "", err
	}

	var info []string
	if properties.Timestamp {
		info = append(info, fmt.Sprintf("Timestamp:%d", message.Timestamp.UnixMilli()))
	}
	if properties.PrintOffset {
		info = append(info, fmt.Sprintf("Partition:%d", message.TopicPartition.Partition))
		info = append(info, fmt.Sprintf("Offset:%s", message.TopicPartition.Offset.String()))
	}
	if len(info) > 0 {
		messageString = fmt.Sprintf("%s\t%s", strings.Join(info, " "), messageString)
	}

	return messageString, nil
}

func (c *command) runConsumer(consumer *ckgo.Consumer, groupHandler *GroupHandler, cmd *cobra.Command) error {
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
				if err := ConsumeMessage(e, groupHandler); err != nil {
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
			case ckgo.OAuthBearerTokenRefresh:
				err := c.refreshOAuthBearerToken(cmd, consumer, e)
				if err != nil {
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

func getFullHeaders(headers []ckgo.Header) []string {
	headerStrings := make([]string, len(headers))
	for i, header := range headers {
		headerStrings[i] = getHeaderString(header)
	}
	return headerStrings
}

func getHeaderString(header ckgo.Header) string {
	if header.Value == nil {
		return fmt.Sprintf("%s=nil", header.Key)
	} else if len(header.Value) == 0 {
		return fmt.Sprintf("%s=<empty>", header.Key)
	} else {
		return fmt.Sprintf(`%s="%s"`, header.Key, string(header.Value))
	}
}

func unmarshalSchemaIdHeader(headers []ckgo.Header) []ckgo.Header {
	for i, header := range headers {
		if header.Key != keySchemaHeaderKey && header.Key != valueSchemaHeaderKey {
			continue
		}

		schemaID := serde.SchemaID{}
		if _, err := schemaID.FromBytes(header.Value); err != nil {
			continue
		}

		headers[i] = ckgo.Header{
			Key:   header.Key,
			Value: []byte(schemaID.GUID.String()),
		}
	}

	return headers
}
