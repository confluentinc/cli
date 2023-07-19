package kafka

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type clientConfig struct {
	language         string // human-friendly language name
	languageId       string // unique id for language used as CLI command
	configId         string // config id used for fetching language config file from the examples Github repo
	isSrApiAvailable bool   // whether SR key pair is supported in the language config file
}

const (
	clientConfigUrlFmt         = "https://raw.githubusercontent.com/confluentinc/examples/master/clients/docs/includes/configs/cloud/%s.config"
	clientConfigDescriptionFmt = "Create a %s client configuration file"

	contextExampleFmt = "confluent kafka client-config create %s"
	flagExampleFmt    = "confluent kafka client-config create %s --environment env-123 --cluster lkc-123456 --api-key my-key --api-secret my-secret"
	srFlagExample     = " --schema-registry-api-key my-sr-key --schema-registry-api-secret my-sr-secret"

	javaConfig         = "java"
	javaSRConfig       = "java-sr"
	librdKafkaConfig   = "librdkafka"
	librdKafkaSRConfig = "librdkafka-sr"
	hoconSRConfig      = "hocon-sr"
	springbootSrConfig = "springboot-sr"
	restproxySrConfig  = "restproxy-sr"

	brokerEndpointTemplate   = "{{ BROKER_ENDPOINT }}"
	clusterApiKeyTemplate    = "{{ CLUSTER_API_KEY }}"
	clusterApiSecretTemplate = "{{ CLUSTER_API_SECRET }}"
	srEndpointTemplate       = "https://{{ SR_ENDPOINT }}"
	srApiKeyTemplate         = "{{ SR_API_KEY }}"
	srApiSecretTemplate      = "{{ SR_API_SECRET }}"

	srEndpointProperty          = "schema.registry.url"
	srCredentialsSourceProperty = "basic.auth.credentials.source"
	srUserInfoProperty          = "basic.auth.user.info"
)

var (
	clientConfigurations = []*clientConfig{
		{"C#", "csharp", librdKafkaConfig, false},
		{"C/C++", "cpp", librdKafkaConfig, false},
		{"Clojure", "clojure", javaConfig, false},
		{"Go", "go", librdKafkaConfig, false},
		{"Groovy", "groovy", javaConfig, false},
		{"Java", "java", javaSRConfig, true},
		{"Kotlin", "kotlin", javaConfig, false},
		{"Ktor", "ktor", hoconSRConfig, true},
		{"Node.js", "nodejs", librdKafkaConfig, false},
		{"Python", "python", librdKafkaSRConfig, true},
		{"REST API", "restapi", restproxySrConfig, true},
		{"Ruby", "ruby", librdKafkaConfig, false},
		{"Rust", "rust", librdKafkaConfig, false},
		{"Scala", "scala", javaConfig, false},
		{"Spring Boot", "springboot", springbootSrConfig, true},
	}

	re = regexp.MustCompile(fmt.Sprintf("%s|%s|%s", srEndpointProperty, srCredentialsSourceProperty, srUserInfoProperty))
)

func (c *clientConfigCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka client configuration file.",
	}

	for _, language := range clientConfigurations {
		cmd.AddCommand(c.newCreateClientCommand(language))
	}

	return cmd
}

func (c *clientConfigCommand) newCreateClientCommand(clientConfig *clientConfig) *cobra.Command {
	clientConfigDescription := fmt.Sprintf(clientConfigDescriptionFmt, clientConfig.language)
	contextExample := fmt.Sprintf(contextExampleFmt, clientConfig.languageId)
	flagExample := fmt.Sprintf(flagExampleFmt, clientConfig.languageId)

	if clientConfig.isSrApiAvailable {
		contextExample += srFlagExample
		flagExample += srFlagExample
	}

	cmd := &cobra.Command{
		Use:   clientConfig.languageId,
		Short: clientConfigDescription + ".",
		Long:  clientConfigDescription + ", of which the client configuration file is printed to stdout and the warnings are printed to stderr. Please see our examples on how to redirect the command output.",
		Args:  cobra.NoArgs,
		RunE:  c.create(clientConfig.configId, clientConfig.isSrApiAvailable),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: clientConfigDescription + ".",
				Code: contextExample,
			},
			examples.Example{
				Text: clientConfigDescription + " with arguments.",
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

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)

	if clientConfig.isSrApiAvailable {
		cmd.Flags().String("schema-registry-api-key", "", "Schema registry API key.")
		cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API key secret.")
	}

	return cmd
}

func (c *clientConfigCommand) create(configId string, srApiAvailable bool) func(cmd *cobra.Command, _ []string) error {
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

		// replace SR_ENDPOINT, SR_API_KEY, and SR_API_SECRET templates if necessary
		if srApiAvailable {
			configFile, err = c.setSchemaRegistryCluster(cmd, configFile)
			if err != nil {
				return err
			}
		}

		// print configuration file to stdout
		output.Println(configFile)
		return nil
	}
}

func (c *clientConfigCommand) setKafkaCluster(cmd *cobra.Command, configFile string) (string, error) {
	// get kafka cluster from context or flags, including key pair
	kafkaCluster, err := c.Config.Context().GetKafkaClusterForCommand()
	if err != nil {
		return "", err
	}

	if err := addApiKeyToCluster(cmd, kafkaCluster); err != nil {
		return "", err
	}

	// Only validate that the key pair matches with the cluster if it's passed via the flag.
	// This is because currently "api-key store" does not check if the secret is valid. Therefore, if users
	// choose to use the key pair stored in the context, we should use it without doing a validation.
	flagKey, flagSecret, err := c.Config.Context().KeyAndSecretFlags(cmd)
	if err != nil {
		return "", err
	}
	if flagKey != "" || flagSecret != "" {
		if err := c.validateKafkaCredentials(kafkaCluster); err != nil {
			return "", err
		}
	} else {
		if err := kafkaCluster.DecryptAPIKeys(); err != nil {
			return "", err
		}
	}

	// replace BROKER_ENDPOINT, CLUSTER_API_KEY, and CLUSTER_API_SECRET templates
	configFile = replaceTemplates(configFile, map[string]string{
		brokerEndpointTemplate:   kafkaCluster.Bootstrap,
		clusterApiKeyTemplate:    kafkaCluster.APIKey,
		clusterApiSecretTemplate: kafkaCluster.GetApiSecret(),
	})
	return configFile, nil
}

func (c *clientConfigCommand) setSchemaRegistryCluster(cmd *cobra.Command, configFile string) (string, error) {
	// get schema registry cluster from context and flags, including key pair
	srCluster, err := c.getSchemaRegistryCluster(cmd)
	if err != nil {
		if err.Error() == errors.NotLoggedInErrorMsg {
			return "", new(errors.SRNotAuthenticatedError)
		}
		// if SR not enabled, comment out SR in the configuration file and warn users
		if srNotEnabledErr, ok := err.(*errors.SRNotEnabledError); ok {
			return commentAndWarnAboutSchemaRegistry(srNotEnabledErr.ErrorMsg, srNotEnabledErr.SuggestionsMsg, configFile), nil
		}
		return "", err
	}

	// replace SR_ENDPOINT template
	configFile = replaceTemplates(configFile, map[string]string{
		srEndpointTemplate: srCluster.SchemaRegistryEndpoint,
	})

	// if empty API key or secret, comment out SR in the configuration file (but still replace SR_ENDPOINT) and warn users
	if len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0 {
		// comment out SR and warn users
		if len(srCluster.SrCredentials.Key) == 0 && len(srCluster.SrCredentials.Secret) == 0 {
			// both key and secret empty
			configFile = commentAndWarnAboutSchemaRegistry(errors.SRCredsNotSetReason, errors.SRCredsNotSetSuggestions, configFile)
		} else if len(srCluster.SrCredentials.Key) == 0 {
			// only key empty
			configFile = commentAndWarnAboutSchemaRegistry(errors.SRKeyNotSetReason, errors.SRKeyNotSetSuggestions, configFile)
		} else {
			// only secret empty
			configFile = commentAndWarnAboutSchemaRegistry(fmt.Sprintf(errors.SRSecretNotSetReason, srCluster.SrCredentials.Key), errors.SRSecretNotSetSuggestions, configFile)
		}

		return configFile, nil
	}

	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return "", err
	}

	// validate that the key pair matches with the cluster
	if err := c.validateSchemaRegistryCredentials(srCluster, unsafeTrace); err != nil {
		return "", err
	}

	// replace SR_API_KEY and SR_API_SECRET templates
	configFile = replaceTemplates(configFile, map[string]string{
		srApiKeyTemplate:    srCluster.SrCredentials.Key,
		srApiSecretTemplate: srCluster.SrCredentials.Secret,
	})
	return configFile, nil
}

// TODO: once dynamic_context::SchemaRegistryCluster consolidates the SR API key stored in the context and
// the key passed via the flags, please remove this function entirely because there is no more need to
// manually fetch the values of the flags. (see setKafkaCluster as example)
func (c *clientConfigCommand) getSchemaRegistryCluster(cmd *cobra.Command) (*v1.SchemaRegistryCluster, error) {
	// get SR cluster from context
	srCluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return nil, err
	}

	// get SR key pair from flag
	schemaRegistryApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
	if err != nil {
		return nil, err
	}
	schemaRegistryApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
	if err != nil {
		return nil, err
	}

	// set SR key pair
	srCluster.SrCredentials = &v1.APIKeyPair{
		Key:    schemaRegistryApiKey,
		Secret: schemaRegistryApiSecret,
	}
	return srCluster, nil
}

func (c *clientConfigCommand) validateKafkaCredentials(kafkaCluster *v1.KafkaClusterConfig) error {
	configMap, err := getCommonConfig(kafkaCluster, c.clientId)
	if err != nil {
		return err
	}
	adminClient, err := ckafka.NewAdminClient(configMap)
	if err != nil {
		return err
	}
	defer adminClient.Close()
	timeout := 5 * time.Second
	if _, err := adminClient.GetMetadata(nil, true, int(timeout.Milliseconds())); err != nil {
		if err.Error() == ckafka.ErrTransport.String() {
			err = errors.NewErrorWithSuggestions(errors.KafkaCredsValidationFailedErrorMsg, errors.KafkaCredsValidationFailedSuggestions)
		}
		return err
	}

	return nil
}

func (c *clientConfigCommand) validateSchemaRegistryCredentials(srCluster *v1.SchemaRegistryCluster, unsafeTrace bool) error {
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
	srConfig.Debug = unsafeTrace
	srClient := srsdk.NewAPIClient(srConfig)

	// Test credentials
	if _, _, err := srClient.DefaultApi.Get(srCtx); err != nil {
		return errors.NewErrorWithSuggestions(errors.SRCredsValidationFailedErrorMsg, errors.SRCredsValidationFailedSuggestions)
	}

	return nil
}

func fetchConfigFile(configId string) (string, error) {
	url := fmt.Sprintf(clientConfigUrlFmt, configId)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf(errors.FetchConfigFileErrorMsg, resp.StatusCode)
	}

	defer resp.Body.Close()

	configFile, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(configFile), nil
}

func replaceTemplates(configFile string, m map[string]string) string {
	for template, value := range m {
		configFile = strings.ReplaceAll(configFile, template, value)
	}
	return configFile
}

func commentAndWarnAboutSchemaRegistry(reason, suggestions, configFile string) string {
	warning := errors.NewWarningWithSuggestions(errors.SRInConfigFileWarning, reason, suggestions+"\n"+errors.SRInConfigFileSuggestions)
	output.ErrPrint(warning.DisplayWarningWithSuggestions())

	return commentSchemaRegistryLines(configFile)
}

func commentSchemaRegistryLines(configFile string) string {
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
	lines := strings.Split(configFile, "\n")

	for idx, line := range lines {
		// if contains one of the SR lines
		if re.MatchString(line) {
			// find the first non-space index in the line -- aka find where to insert #
			firstNonSpaceIdx := strings.IndexFunc(line, func(c rune) bool {
				return !unicode.IsSpace(c)
			})
			// insert #
			lines[idx] = line[:firstNonSpaceIdx] + "#" + line[firstNonSpaceIdx:]
		}
	}

	return strings.Join(lines, "\n")
}
