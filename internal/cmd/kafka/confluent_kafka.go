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
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
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

// on-prem authentication

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
			err = fmt.Errorf("token retrieval error: %v", retrieveErr)
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
				err = fmt.Errorf("error setting token and extensions: %v", setTokenError)
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

// create cloud kafka client

func newProducer(kafka *configv1.KafkaClusterConfig, clientID, configPath string, configStrings []string) (*ckafka.Producer, error) {
	configMap, err := getProducerConfigMap(kafka, clientID)
	if err != nil {
		return nil, err
	}

	return newProducerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newConsumer(group string, kafka *configv1.KafkaClusterConfig, clientID string, beginning bool, configPath string, configStrings []string) (*ckafka.Consumer, error) {
	configMap, err := getConsumerConfigMap(group, kafka, clientID, beginning)
	if err != nil {
		return nil, err
	}

	return newConsumerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

// create on-prem kafka client

func newOnPremProducer(cmd *cobra.Command, clientID string, configPath string, configStrings []string) (*ckafka.Producer, error) {
	configMap, err := getOnPremProducerConfigMap(cmd, clientID)
	if err != nil {
		return nil, err
	}

	return newProducerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

func newOnPremConsumer(cmd *cobra.Command, clientID string, configPath string, configStrings []string) (*ckafka.Consumer, error) {
	configMap, err := getOnPremConsumerConfigMap(cmd, clientID)
	if err != nil {
		return nil, err
	}

	return newConsumerWithOverwrittenConfigs(configMap, configPath, configStrings)
}

// consumer utilities

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

// schema utilities

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
