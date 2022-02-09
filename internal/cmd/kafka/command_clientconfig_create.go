package kafka

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"unicode"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

type createCommand struct {
	*pcmd.HasAPIKeyCLICommand
	prerunner pcmd.PreRunner
	clientId  string
}

type clientConfig struct {
	language       string // human-friendly language name
	languageId     string // unique id for language used as CLI command
	configId       string // config id used for fetching language config file from he examples github repo
	srApiAvailable bool   // whether SR key pair is supported in the language config file
}

const (
	clientConfigUrlFmt         = "https://raw.githubusercontent.com/confluentinc/examples/master/clients/docs/includes/configs/cloud/%s.config"
	clientConfigDescriptionFmt = "Create a %s Client configuration file"

	contextExampleFmt = "confluent kafka client-config create %s"
	flagExampleFmt    = "confluent kafka client-config create %s --environment env-123 --cluster lkc-123456 --api-key my-key --api-secret my-secret"
	srFlagExample     = " --sr-apikey my-sr-key --sr-apisecret my-sr-secret"

	javaConfig         = "java"
	javaSRConfig       = "java-sr"
	librdKafkaConfig   = "librdkafka"
	librdKafkaSRConfig = "librdkafka-sr"
	hoconSRConfig      = "hocon-sr"
	springbootSrConfig = "springboot-sr"
	restproxySrConfig  = "restproxy-sr"

	brokerEndpointTemplate       = "{{ BROKER_ENDPOINT }}"
	clusterApiKeyTemplate        = "{{ CLUSTER_API_KEY }}"
	clusterApiSecretTemplate     = "{{ CLUSTER_API_SECRET }}"
	srEndpointTemplate           = "https://{{ SR_ENDPOINT }}"
	srApiKeyTemplate             = "{{ SR_API_KEY }}"
	srApiSecretTemplate          = "{{ SR_API_SECRET }}"
	srBasicAuthCredentialsSource = "USER_INFO"
)

var (
	clojure    = &clientConfig{"Clojure", "clojure", javaConfig, false}
	cpp        = &clientConfig{"C/C++", "cpp", librdKafkaConfig, false}
	csharp     = &clientConfig{"C#", "csharp", librdKafkaConfig, false}
	golang     = &clientConfig{"Go", "go", librdKafkaConfig, false}
	groovy     = &clientConfig{"Groovy", "groovy", javaConfig, false}
	java       = &clientConfig{"Java", "java", javaSRConfig, true}
	kotlin     = &clientConfig{"Kotlin", "kotlin", javaConfig, false}
	ktor       = &clientConfig{"Ktor", "ktor", hoconSRConfig, true}
	nodeJS     = &clientConfig{"Node.js", "nodejs", librdKafkaConfig, false}
	python     = &clientConfig{"Python", "python", librdKafkaSRConfig, true}
	restAPI    = &clientConfig{"REST API", "restapi", restproxySrConfig, true}
	ruby       = &clientConfig{"Ruby", "ruby", librdKafkaConfig, false}
	rust       = &clientConfig{"Rust", "rust", librdKafkaConfig, false}
	scala      = &clientConfig{"Scala", "scala", javaConfig, false}
	springBoot = &clientConfig{"Spring Boot", "springboot", springbootSrConfig, true}

	languages = []*clientConfig{
		clojure, cpp, csharp, golang, groovy, java, kotlin, ktor, nodeJS, python, ruby, rust, scala, springBoot, restAPI}
)

func (c *clientConfigCommand) newCreateCommand() *cobra.Command {
	cliCmd := pcmd.NewHasAPIKeyCLICommand(
		&cobra.Command{
			Use:         "create",
			Short:       "Create a Kafka Client configuration file.",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		}, c.prerunner)

	cmd := &createCommand{
		HasAPIKeyCLICommand: cliCmd,
		prerunner:           c.prerunner,
		clientId:            c.clientId,
	}

	for _, language := range languages {
		cmd.addCommand(language)
	}

	return cmd.Command
}

func (c *createCommand) addCommand(clientConfig *clientConfig) {
	var clientConfigDescription, contextExample, flagExample string

	clientConfigDescription = fmt.Sprintf(clientConfigDescriptionFmt, clientConfig.language)
	contextExample = fmt.Sprintf(contextExampleFmt, clientConfig.languageId)
	flagExample = fmt.Sprintf(flagExampleFmt, clientConfig.languageId)

	if clientConfig.srApiAvailable {
		contextExample += srFlagExample
		flagExample += srFlagExample
	}

	cmd := &cobra.Command{
		Use:   clientConfig.languageId,
		Short: clientConfigDescription + ".",
		Long: clientConfigDescription + ", of which the client configuration file is printed to stdout and " +
			"the warnings is printed to stderr. Please see our examples on how to redirect the command output.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.create(clientConfig.configId, clientConfig.srApiAvailable)),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: clientConfigDescription + ".",
				Code: contextExample,
			},
			examples.Example{
				Text: clientConfigDescription + " with arguments passed via flags.",
				Code: flagExample,
			},
			examples.Example{
				Text: clientConfigDescription + ", redirecting the configuration to a file and the warnings to a separate file.",
				Code: contextExample + " 1> my-client-config-file.config 2> my-warnings-file",
			},
			examples.Example{
				Text: clientConfigDescription + ", redirecting the configuration to a file and keeping the warnings in the console.",
				Code: contextExample + " 1> my-client-config-file.config 2>&1",
			},
		),
	}

	//add authenticated state flags
	pcmd.AddContextFlag(cmd, c.CLICommand)
	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")

	//add api key flags
	cmd.Flags().String("api-key", "", "API key.")
	pcmd.AddApiSecretFlag(cmd)

	// add sr flags
	if clientConfig.srApiAvailable {
		cmd.Flags().String("sr-apikey", "", "Schema registry API key.")
		cmd.Flags().String("sr-apisecret", "", "Schema registry API key secret.")
	}

	c.AddCommand(cmd)
}

func (c *createCommand) create(configId string, srApiAvailable bool) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		// fetch raw configuration file in which templates need to be replaced
		configFile, err := fetchConfigFile(configId)
		if err != nil {
			return err
		}

		// replace BROKER_ENDPOINT, CLUSTER_API_KEY, and CLUSTER_API_SECRET templates
		configFile, err = c.setKafkaCluster(cmd, configFile)
		if err != nil {
			return err
		}

		// replace SR_ENDPOINT, SR_API_KEY, and SR_API_SECRET templates
		configFile, err = c.setSchemaRegistryCluster(cmd, configFile, srApiAvailable)
		if err != nil {
			return err
		}

		// print configuration file to stdout
		utils.Println(cmd, string(configFile))
		return nil
	}
}

func (c *createCommand) setKafkaCluster(cmd *cobra.Command, configFile []byte) ([]byte, error) {
	// get kafka cluster from context or flags, including key pair
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	// only validate that the key pair matches with the cluster if it's passed via the flag.
	// this is because currently "api-key store" does not check if the secret is valid. therefore, if users
	// choose to use the key pair stored in the context, we should use it without doing a validation.
	// TODO: always validate key pair after feature enhancement: https://confluentinc.atlassian.net/browse/CLI-1575
	flagKey, flagSecret, err := c.Context.KeyAndSecretFlags(cmd)
	if err != nil {
		return nil, err
	}
	if len(flagKey) != 0 && len(flagSecret) != 0 {
		err = c.validateKafkaCredentials(kafkaCluster)
		if err != nil {
			return nil, err
		}
	}

	// replace BROKER_ENDPOINT, CLUSTER_API_KEY, and CLUSTER_API_SECRET templates
	configFile = bytes.ReplaceAll(configFile, []byte(brokerEndpointTemplate), []byte(kafkaCluster.Bootstrap))
	configFile = bytes.ReplaceAll(configFile, []byte(clusterApiKeyTemplate), []byte(kafkaCluster.APIKey))
	configFile = bytes.ReplaceAll(configFile, []byte(clusterApiSecretTemplate), []byte(kafkaCluster.APIKeys[kafkaCluster.APIKey].Secret))
	return configFile, nil
}

func (c *createCommand) setSchemaRegistryCluster(cmd *cobra.Command, configFile []byte, srApiAvailable bool) ([]byte, error) {
	// the language does not support SR so no need to modify configuration file. return directly
	if !srApiAvailable {
		return configFile, nil
	}

	// get schema registry cluster from context and flags, including key pair
	srCluster, err := c.getSchemaRegistryCluster(cmd)
	if err != nil {
		// if SR not enabled, comment out SR in the configuration file and warn users
		if errors.CatchSchemaRegistryNotEnabledError(err) {
			return commentAndWarnAboutSr(errors.SRNotEnabledErrorMsg, errors.SRNotEnabledSuggestions, configFile)
		}
		return nil, err
	}

	// if empty api key or secret, comment out SR in the configuration file (but still replace SR_ENDPOINT) and warn users
	if len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0 {
		// comment out SR and warn users
		if len(srCluster.SrCredentials.Key) == 0 && len(srCluster.SrCredentials.Secret) == 0 {
			// both key and secret empty
			configFile, err = commentAndWarnAboutSr(errors.SRCredsNotSetReason, errors.SRCredsNotSetSuggestions, configFile)
			if err != nil {
				return nil, err
			}
		} else if len(srCluster.SrCredentials.Key) == 0 {
			// only key empty
			configFile, err = commentAndWarnAboutSr(errors.SRKeyNotSetReason, errors.SRKeyNotSetSuggestions, configFile)
			if err != nil {
				return nil, err
			}
		} else {
			// only secret empty
			configFile, err = commentAndWarnAboutSr(fmt.Sprintf(errors.SRSecretNotSetReason, srCluster.SrCredentials.Key), errors.SRSecretNotSetSuggestions, configFile)
			if err != nil {
				return nil, err
			}
		}

		// replace SR_ENDPOINT template
		configFile = bytes.ReplaceAll(configFile, []byte(srEndpointTemplate), []byte(srCluster.SchemaRegistryEndpoint))
		return configFile, nil
	}

	// validate that the key pair matches with the cluster
	err = c.validateSchemaRegistryCredentials(srCluster)
	if err != nil {
		if err.Error() == "ccloud" {
			return nil, new(errors.SRNotAuthenticatedError)
		} else {
			return nil, err
		}
	}

	// replace SR_ENDPOINT, SR_API_KEY and SR_API_SECRET templates
	configFile = bytes.ReplaceAll(configFile, []byte(srEndpointTemplate), []byte(srCluster.SchemaRegistryEndpoint))
	configFile = bytes.ReplaceAll(configFile, []byte(srApiKeyTemplate), []byte(srCluster.SrCredentials.Key))
	configFile = bytes.ReplaceAll(configFile, []byte(srApiSecretTemplate), []byte(srCluster.SrCredentials.Secret))
	return configFile, nil
}

func (c *createCommand) getSchemaRegistryCluster(cmd *cobra.Command) (*v1.SchemaRegistryCluster, error) {
	// get SR cluster from context
	srCluster, err := c.Context.SchemaRegistryCluster(cmd)
	if err != nil {
		return nil, err
	}

	// get SR key pair from flag
	apiKey, err := cmd.Flags().GetString("sr-apikey")
	if err != nil {
		return nil, err
	}
	apiSecret, err := cmd.Flags().GetString("sr-apisecret")
	if err != nil {
		return nil, err
	}

	// set SR key pair
	srCluster.SrCredentials = &v1.APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
	return srCluster, nil
}

func (c *createCommand) validateKafkaCredentials(kafkaCluster *v1.KafkaClusterConfig) error {
	// validate
	adminClient, err := ckafka.NewAdminClient(getCommonConfig(kafkaCluster, c.clientId))
	if err != nil {
		return err
	}
	defer adminClient.Close()
	timeout := 5 * time.Second
	_, err = adminClient.GetMetadata(nil, true, int(timeout.Milliseconds()))
	if err != nil {
		if err.Error() == ckafka.ErrTransport.String() {
			err = errors.NewErrorWithSuggestions(errors.KafkaCredsValidationFailedErrorMsg, errors.KafkaCredsValidationFailedSuggestions)
		}
		return err
	}

	return nil
}

func (c *createCommand) validateSchemaRegistryCredentials(srCluster *v1.SchemaRegistryCluster) error {
	srConfig := srsdk.NewConfiguration()

	// set BasePath of srConfig
	srConfig.BasePath = srCluster.SchemaRegistryEndpoint

	// get credentials as SR basic auth
	srAuth := &srsdk.BasicAuth{}
	if srCluster.SrCredentials != nil {
		srAuth.UserName = srCluster.SrCredentials.Key
		srAuth.Password = srCluster.SrCredentials.Secret
	}
	srCtx := context.WithValue(context.Background(), srsdk.ContextBasicAuth, *srAuth)

	srConfig.UserAgent = c.Version.UserAgent
	srClient := srsdk.NewAPIClient(srConfig)

	// Test credentials
	if _, _, err := srClient.DefaultApi.Get(srCtx); err != nil {
		return errors.NewErrorWithSuggestions(errors.SRCredsValidationFailedErrorMsg, errors.SRCredsValidationFailedSuggestions)
	}

	return nil
}

func fetchConfigFile(configId string) ([]byte, error) {
	url := fmt.Sprintf(clientConfigUrlFmt, configId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(errors.FetchConfigFileErrorMsg, resp.StatusCode)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func commentAndWarnAboutSr(reason string, suggestions string, configFile []byte) ([]byte, error) {
	warning := errors.NewWarningWithSuggestions(
		errors.SRInConfigFileWarning,
		reason,
		suggestions+"\n"+errors.SRInConfigFileSuggestions)
	warning.DisplayWarningWithSuggestions()

	configFile, err := comment(configFile, srEndpointTemplate)
	if err != nil {
		return nil, err
	}
	configFile, err = comment(configFile, srBasicAuthCredentialsSource)
	if err != nil {
		return nil, err
	}
	configFile, err = comment(configFile, srApiKeyTemplate)
	if err != nil {
		return nil, err
	}

	return configFile, nil
}

func comment(configFile []byte, template string) ([]byte, error) {
	/* Examples:
	1. Case where SR properties start at the beginning of the line
	# Required connection configs for Confluent Cloud Schema Registry
	schema.registry.url=https://{{ SR_ENDPOINT }}
	basic.auth.credentials.source=USER_INFO
	basic.auth.user.info={{ SR_API_KEY }}:{{ SR_API_SECRET }}

	---BECOMES--->

	# Required connection configs for Confluent Cloud Schema Registry
	#schema.registry.url=https://{{ SR_ENDPOINT }}
	#basic.auth.credentials.source=USER_INFO
	#basic.auth.user.info={{ SR_API_KEY }}:{{ SR_API_SECRET }}

	2. Case where SR properties don't start at the beginning of the line
	properties {
		# Required connection configs for Confluent Cloud Schema Registry
		schema.registry.url = "https://{{ SR_ENDPOINT }}"
		basic.auth.credentials.source = USER_INFO
		basic.auth.user.info = "{{ SR_API_KEY }}:{{ SR_API_SECRET }}"
	}

	---BECOMES--->

	properties {
		# Required connection configs for Confluent Cloud Schema Registry
		#schema.registry.url = "https://{{ SR_ENDPOINT }}"
		#basic.auth.credentials.source = USER_INFO
		#basic.auth.user.info = "{{ SR_API_KEY }}:{{ SR_API_SECRET }}"
	}
	*/

	// find the index of the template -- aka find the line of the template
	endpointIdx := bytes.Index(configFile, []byte(template))

	// find the last new line index before the template -- aka find the beginning of the template line
	lastNewLineIdx := bytes.LastIndex(configFile[:endpointIdx], []byte("\n"))

	// sanity check
	if lastNewLineIdx == -1 || lastNewLineIdx+1 >= endpointIdx {
		return nil, errors.New(errors.WriteConfigFileErrorMsg)
	}

	// find the first non-space index in the template line -- aka find where to insert #
	fistNonSpaceIdx := bytes.IndexFunc(configFile[lastNewLineIdx:], func(c rune) bool {
		return !unicode.IsSpace(c)
	})

	// insert #
	configFile = insertByte(configFile, '#', lastNewLineIdx+fistNonSpaceIdx)

	return configFile, nil
}

func insertByte(arr []byte, element byte, idx int) []byte {
	arr = append(arr, 0)
	copy(arr[idx+1:], arr[idx:])
	arr[idx] = element
	return arr
}
