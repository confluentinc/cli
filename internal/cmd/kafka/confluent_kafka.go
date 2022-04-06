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

	"github.com/antihax/optional"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	serdes "github.com/confluentinc/cli/internal/pkg/serdes"
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
	Subject    string
	Properties ConsumerProperties
}

func (c *authenticatedTopicCommand) refreshOAuthBearerToken(cmd *cobra.Command, client ckafka.Handle) error {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return err
	}
	mechanism, err := cmd.Flags().GetString("sasl-mechanism")
	if err != nil {
		return err
	}
	if protocol == "SASL_SSL" && mechanism == "OAUTHBEARER" {
		oart := ckafka.OAuthBearerTokenRefresh{Config: oauthConfig}
		if c.State == nil { // require log-in to use oauthbearer token
			return errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestion)
		}
		oauthBearerToken, retrieveErr := retrieveUnsecuredToken(oart, c.AuthToken())
		if retrieveErr != nil {
			err = fmt.Errorf("Token retrieval error: %v\n", retrieveErr)
			if err != nil {
				return err
			}
			err = client.SetOAuthBearerTokenFailure(retrieveErr.Error())
			if err != nil {
				return err
			}
		} else {
			setTokenError := client.SetOAuthBearerToken(oauthBearerToken)
			if setTokenError != nil {
				err = fmt.Errorf("Error setting token and extensions: %v\n", setTokenError)
				if err != nil {
					return err
				}
				err = client.SetOAuthBearerTokenFailure(setTokenError.Error())
				if err != nil {
					return err
				}
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
	configMap := &ckafka.ConfigMap{
		"security.protocol":                     "SASL_SSL",
		"sasl.mechanism":                        "PLAIN",
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     kafka.Bootstrap,
		"sasl.username":                         kafka.APIKey,
		"sasl.password":                         kafka.APIKeys[kafka.APIKey].Secret,
	}
	return configMap
}

func getOnPremProducerConfigMap(cmd *cobra.Command, clientID string) (*ckafka.ConfigMap, error) {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
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
		"enable.ssl.certificate.verification":   true,
		"ssl.ca.location":                       caLocation,
		"retry.backoff.ms":                      "250",
		"request.timeout.ms":                    "10000",
	}
	if err := setProducerDebugOption(configMap); err != nil {
		return nil, err
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
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return nil, err
	}
	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"group.id":                              group,
		"bootstrap.servers":                     bootstrap,
		"enable.ssl.certificate.verification":   true,
		"ssl.ca.location":                       caLocation,
	}
	autoOffsetReset := "latest"
	if beginning {
		autoOffsetReset = "earliest"
	}
	if err := configMap.SetKey("auto.offset.reset", autoOffsetReset); err != nil {
		return nil, err
	}
	if err := setConsumerDebugOption(configMap); err != nil {
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
	default:
		return nil, errors.NewErrorWithSuggestions(fmt.Errorf(errors.InvalidSecurityProtocolErrorMsg, protocol).Error(), errors.OnPremConfigGuideSuggestion)
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
	mechanism, err := cmd.Flags().GetString("sasl-mechanism")
	if err != nil {
		return nil, err
	}
	saslMap := map[string]string{
		"security.protocol": "SASL_SSL",
		"sasl.mechanism":    mechanism,
	}
	if mechanism == "PLAIN" {
		username, password, err := promptForSASLAuth(cmd)
		if err != nil {
			return nil, err
		}
		saslMap["sasl.username"] = username
		saslMap["sasl.password"] = password
	} else {
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
	if err := setProducerDebugOption(configMap); err != nil {
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
	if err := setConsumerDebugOption(configMap); err != nil {
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
	tempStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%s-%d.txt", h.Subject, schemaID))
	tempRefStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%s-%d.ref", h.Subject, schemaID))
	var references []srsdk.SchemaReference
	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		getSchemaOpts := srsdk.GetSchemaOpts{
			Subject: optional.NewString(h.Subject),
		}
		schemaString, _, err := h.SrClient.DefaultApi.GetSchema(h.Ctx, schemaID, &getSchemaOpts)
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
	referencePathMap, err := sr.StoreSchemaReferences(references, h.SrClient, h.Ctx)
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

func promptForSASLAuth(cmd *cobra.Command) (string, string, error) {
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return "", "", err
	}
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return "", "", err
	}
	if username != "" && password != "" {
		return username, password, nil
	}
	f := form.New(
		form.Field{ID: "username", Prompt: "Enter your SASL username"},
		form.Field{ID: "password", Prompt: "Enter your SASL password", IsHidden: true},
	)
	if err := f.Prompt(cmd, form.NewPrompt(os.Stdin)); err != nil {
		return "", "", err
	}
	return f.Responses["username"].(string), f.Responses["password"].(string), nil
}

func setProducerDebugOption(configMap *ckafka.ConfigMap) error {
	switch log.CliLogger.GetLevel() {
	case log.DEBUG:
		return configMap.Set("debug=broker, topic, msg, protocol")
	case log.TRACE:
		return configMap.Set("debug=all")
	}
	return nil
}

func setConsumerDebugOption(configMap *ckafka.ConfigMap) error {
	switch log.CliLogger.GetLevel() {
	case log.DEBUG:
		return configMap.Set("debug=broker, topic, msg, protocol, consumer, cgrp, fetch")
	case log.TRACE:
		return configMap.Set("debug=all")
	}
	return nil
}
