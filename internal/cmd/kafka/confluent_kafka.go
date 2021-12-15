package kafka

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"time"

	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	serdes "github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	messageOffset         = 5 // Schema ID is stored at the [1:5] bytes of a message as meta info (when valid)
	principalClaimNameKey = "principalClaimName"
	principalKey          = "principal"
	joseHeaderEncoded     = "eyJhbGciOiJub25lIn0" // {"alg":"none"}
)

var (
	// Regex for sasl.oauthbearer.config, which constrains it to be
	// 1 or more name=value pairs with optional ignored whitespace
	oauthbearerConfigRegex = regexp.MustCompile("^(\\s*(\\w+)\\s*=\\s*(\\w+))+\\s*$")
	// Regex used to extract name=value pairs from sasl.oauthbearer.config
	oauthbearerNameEqualsValueRegex = regexp.MustCompile("(\\w+)\\s*=\\s*(\\w+)")
)

type ConsumerProperties struct {
	PrintKey   bool
	Delimiter  string
	SchemaPath string
}

// GroupHandler instances are used to handle individual topic-partition claims.
type GroupHandler struct {
	SrClient   *srsdk.APIClient
	Ctx        context.Context
	Format     string
	Out        io.Writer
	Properties ConsumerProperties
}

func refreshOAuthBearerToken(cmd *cobra.Command, client ckafka.Handle, tokenValue string) error {
	mechanism, err := cmd.Flags().GetString("mechanism")
	if err != nil {
		return err
	}
	if mechanism == "OAUTHBEARER" {
		oauthConfig, err := cmd.Flags().GetString("oauthConfig")
		if err != nil {
			return err
		}
		oart := ckafka.OAuthBearerTokenRefresh{Config: oauthConfig}
		oauthBearerToken, retrieveErr := retrieveUnsecuredToken(oart, tokenValue)
		if retrieveErr != nil {
			fmt.Fprintf(os.Stderr, "%% Token retrieval error: %v\n", retrieveErr)
			client.SetOAuthBearerTokenFailure(retrieveErr.Error())
		} else {
			setTokenError := client.SetOAuthBearerToken(oauthBearerToken)
			if setTokenError != nil {
				fmt.Fprintf(os.Stderr, "%% Error setting token and extensions: %v\n", setTokenError)
				client.SetOAuthBearerTokenFailure(setTokenError.Error())
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
	expiration := now.Add(time.Second * time.Duration(600))
	oauthBearerToken := ckafka.OAuthBearerToken{
		TokenValue: tokenValue,
		Expiration: expiration,
		Principal:  principal,
	}
	return oauthBearerToken, nil
}

func NewProducer(kafka *configv1.KafkaClusterConfig, clientID string) (*ckafka.Producer, error) {
	configMap, err := getProducerConfigMap(kafka, clientID)
	if err != nil {
		return nil, err
	}
	return ckafka.NewProducer(configMap)
}

// NewConsumer returns a ConsumerGroup configured for the CLI config
func NewConsumer(group string, kafka *configv1.KafkaClusterConfig, clientID string, beginning bool) (*ckafka.Consumer, error) {
	configMap, err := getConsumerConfigMap(group, kafka, clientID, beginning)
	if err != nil {
		return nil, err
	}
	return ckafka.NewConsumer(configMap)
}

func getCommonConfig(kafka *configv1.KafkaClusterConfig, clientID string) *ckafka.ConfigMap {
	return &ckafka.ConfigMap{
		"security.protocol":                     "SASL_SSL",
		"sasl.mechanism":                        "PLAIN",
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     kafka.Bootstrap,
		"sasl.username":                         kafka.APIKey,
		"sasl.password":                         kafka.APIKeys[kafka.APIKey].Secret,
	}
}

func getOnPremProducerConfigMap(cmd *cobra.Command, clientID string) (*ckafka.ConfigMap, error) {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return nil, err
	}
	enableSSLVerification, err := cmd.Flags().GetBool("ssl-verification")
	if err != nil {
		return nil, err
	}
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return nil, err
	}
	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     bootstrap,
		"enable.ssl.certificate.verification":   enableSSLVerification,
		"ssl.ca.location":                       caLocation,
		"retry.backoff.ms":                      "250",
		"request.timeout.ms":                    "10000",
	}
	return setProtocolConfig(cmd, configMap)
}

func getOnPremConsumerConfigMap(cmd *cobra.Command, clientID string) (*ckafka.ConfigMap, error) {
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}
	if group == "" {
		group = fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	}
	beginning, err := cmd.Flags().GetBool("from-beginning")
	if err != nil {
		return nil, err
	}
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return nil, err
	}
	enableSSLVerification, err := cmd.Flags().GetBool("ssl-verification")
	if err != nil {
		return nil, err
	}
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return nil, err
	}
	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"group.id":                              group,
		"bootstrap.servers":                     bootstrap,
		"enable.ssl.certificate.verification":   enableSSLVerification,
		"ssl.ca.location":                       caLocation,
	}
	autoOffsetReset := "latest"
	if beginning {
		autoOffsetReset = "earliest"
	}
	if err := configMap.SetKey("auto.offset.reset", autoOffsetReset); err != nil {
		return nil, err
	}
	return setProtocolConfig(cmd, configMap)
}

func setProtocolConfig(cmd *cobra.Command, configMap *ckafka.ConfigMap) (*ckafka.ConfigMap, error) {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return nil, err
	}
	switch protocol {
	case "SSL":
		configMap, err = setSSLConfig(cmd, configMap)
		if err != nil {
			return nil, err
		}
	case "SASL_SSL":
		configMap, err = setSASLConfig(cmd, configMap)
		if err != nil {
			return nil, err
		}
	}
	return configMap, nil
}

func setSSLConfig(cmd *cobra.Command, configMap *ckafka.ConfigMap) (*ckafka.ConfigMap, error) {
	certLocation, err := cmd.Flags().GetString("cert-location")
	if err != nil {
		return nil, err
	}
	keyLocation, err := cmd.Flags().GetString("key-location")
	if err != nil {
		return nil, err
	}
	keyPassword, err := cmd.Flags().GetString("key-password")
	if err != nil {
		return nil, err
	}
	sslMap := map[string]string{
		"security.protocol":        "SSL",
		"ssl.certificate.location": certLocation,
		"ssl.key.location":         keyLocation,
		"ssl.key.password":         keyPassword,
	}
	for key, value := range sslMap {
		if err := configMap.SetKey(key, value); err != nil {
			return nil, err
		}
	}
	return configMap, nil
}

func setSASLConfig(cmd *cobra.Command, configMap *ckafka.ConfigMap) (*ckafka.ConfigMap, error) {
	mechanism, err := cmd.Flags().GetString("mechanism")
	if err != nil {
		return nil, err
	}
	saslMap := map[string]string{
		"security.protocol": "SASL_SSL",
		"sasl.mechanism":    mechanism,
	}
	if mechanism == "PLAIN" {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return nil, err
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return nil, err
		}
		saslMap["sasl.username"] = username
		saslMap["sasl.password"] = password
	} else {
		oauthConfig, err := cmd.Flags().GetString("oauthConfig")
		if err != nil {
			return nil, err
		}
		saslMap["sasl.oauthbearer.config"] = oauthConfig
	}
	for key, value := range saslMap {
		if err := configMap.SetKey(key, value); err != nil {
			return nil, err
		}
	}
	return configMap, nil
}

func getProducerConfigMap(kafka *configv1.KafkaClusterConfig, clientID string) (*ckafka.ConfigMap, error) {
	configMap := getCommonConfig(kafka, clientID)
	if err := configMap.SetKey("retry.backoff.ms", "250"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("request.timeout.ms", "10000"); err != nil {
		return nil, err
	}
	return configMap, nil
}

func getConsumerConfigMap(group string, kafka *configv1.KafkaClusterConfig, clientID string, beginning bool) (*ckafka.ConfigMap, error) {
	configMap := getCommonConfig(kafka, clientID)
	if err := configMap.SetKey("group.id", group); err != nil {
		return nil, err
	}
	autoOffsetReset := "latest"
	if beginning {
		autoOffsetReset = "earliest"
	}
	if err := configMap.SetKey("auto.offset.reset", autoOffsetReset); err != nil {
		return nil, err
	}
	return configMap, nil
}

func (h *GroupHandler) RequestSchema(value []byte) (string, map[string]string, error) {
	if len(value) < messageOffset {
		return "", nil, errors.New(errors.FailedToFindSchemaIDErrorMsg)
	}

	// Retrieve schema from cluster only if schema is specified.
	schemaID := int32(binary.BigEndian.Uint32(value[1:messageOffset])) // schema id is stored as a part of message meta info

	// Create temporary file to store schema retrieved (also for cache). Retry if get error retriving schema or writing temp schema file
	tempStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%d.txt", schemaID))
	tempRefStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%d.ref", schemaID))
	var references []srsdk.SchemaReference
	if !fileExists(tempStorePath) || !fileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		schemaString, _, err := h.SrClient.DefaultApi.GetSchema(h.Ctx, schemaID, nil)
		if err != nil {
			return "", nil, err
		}
		err = ioutil.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644)
		if err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		err = ioutil.WriteFile(tempRefStorePath, refBytes, 0644)
		if err != nil {
			return "", nil, err
		}
		references = schemaString.References
	} else {
		refBlob, err := ioutil.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		err = json.Unmarshal(refBlob, &references)
		if err != nil {
			return "", nil, err
		}
	}

	// Store the references in temporary files
	referencePathMap, err := storeSchemaReferences(references, h.SrClient, h.Ctx)
	if err != nil {
		return "", nil, err
	}

	return tempStorePath, referencePathMap, nil
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
		_, err := fmt.Fprint(h.Out, keyString+h.Properties.Delimiter)
		if err != nil {
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
		err = deserializationProvider.LoadSchema(schemaPath, referencePathMap)
		if err != nil {
			return err
		}
	}
	jsonMessage, err := serdes.Deserialize(deserializationProvider, value)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(h.Out, jsonMessage)
	if err != nil {
		return err
	}

	if e.Headers != nil {
		_, err = fmt.Fprintf(h.Out, "%% Headers: %v\n", e.Headers)
		if err != nil {
			return err
		}
	}
	return nil
}

func runConsumer(cmd *cobra.Command, consumer *ckafka.Consumer, groupHandler *GroupHandler) error {
	run := true
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for run {
		select {
		case <-signals: // Trap SIGINT to trigger a shutdown.
			utils.ErrPrintln(cmd, errors.StoppingConsumer)
			consumer.Close()
			run = false
		default:
			event := consumer.Poll(100) // polling event from consumer with a timeout of 100ms
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *ckafka.Message:
				err := consumeMessage(e, groupHandler)
				if err != nil {
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
