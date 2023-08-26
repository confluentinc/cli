package kafka

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/color"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	schemaregistry "github.com/confluentinc/cli/v3/pkg/schema-registry"
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
		cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API secret.")
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
		color.Println(c.Config.EnableColor, configFile)
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
	flagKey, _, err := c.Config.Context().KeyAndSecretFlags(cmd)
	if err != nil {
		return "", err
	}
	if flagKey != "" {
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
	cluster, err := c.getSchemaRegistryCluster()
	if err != nil {
		return "", err
	}

	schemaRegistryApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
	if err != nil {
		return "", err
	}

	schemaRegistryApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
	if err != nil {
		return "", err
	}

	apiKeyPair := &config.APIKeyPair{
		Key:    schemaRegistryApiKey,
		Secret: schemaRegistryApiSecret,
	}

	// replace SR_ENDPOINT template
	configFile = replaceTemplates(configFile, map[string]string{
		srEndpointTemplate: cluster.Spec.GetHttpEndpoint(),
	})

	// if empty API key or secret, comment out SR in the configuration file (but still replace SR_ENDPOINT) and warn users
	if apiKeyPair.Key == "" || apiKeyPair.Secret == "" {
		// comment out SR and warn users
		if apiKeyPair.Key == "" && apiKeyPair.Secret == "" {
			// both key and secret empty
			configFile = commentAndWarnAboutSchemaRegistry(errors.SRCredsNotSetReason, errors.SRCredsNotSetSuggestions, configFile)
		} else if apiKeyPair.Key == "" {
			// only key empty
			configFile = commentAndWarnAboutSchemaRegistry(errors.SRKeyNotSetReason, errors.SRKeyNotSetSuggestions, configFile)
		} else {
			// only secret empty
			configFile = commentAndWarnAboutSchemaRegistry(fmt.Sprintf(errors.SRSecretNotSetReason, apiKeyPair.Key), errors.SRSecretNotSetSuggestions, configFile)
		}

		return configFile, nil
	}

	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return "", err
	}

	// validate that the key pair matches with the cluster
	if err := c.validateSchemaRegistryCredentials(cluster, apiKeyPair, unsafeTrace); err != nil {
		return "", err
	}

	// replace SR_API_KEY and SR_API_SECRET templates
	configFile = replaceTemplates(configFile, map[string]string{
		srApiKeyTemplate:    apiKeyPair.Key,
		srApiSecretTemplate: apiKeyPair.Secret,
	})
	return configFile, nil
}

func (c *clientConfigCommand) getSchemaRegistryCluster() (*srcmv2.SrcmV2Cluster, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return nil, errors.NewSRNotEnabledError()
	}

	return &clusters[0], nil
}

func (c *clientConfigCommand) validateKafkaCredentials(kafkaCluster *config.KafkaClusterConfig) error {
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

func (c *clientConfigCommand) validateSchemaRegistryCredentials(cluster *srcmv2.SrcmV2Cluster, apiKeyPair *config.APIKeyPair, unsafeTrace bool) error {
	srConfig := srsdk.NewConfiguration()
	srConfig.BasePath = cluster.Spec.GetHttpEndpoint()
	srConfig.UserAgent = c.Version.UserAgent
	srConfig.Debug = unsafeTrace

	client := schemaregistry.NewClientWithApiKey(srConfig, apiKeyPair.Key, apiKeyPair.Secret)

	if err := client.Get(); err != nil {
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
