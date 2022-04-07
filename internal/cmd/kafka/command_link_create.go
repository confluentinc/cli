package kafka

import (
	"fmt"

	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	sourceApiKeyFlagName               = "source-api-key"
	sourceApiSecretFlagName            = "source-api-secret"
	destinationApiKeyFlagName          = "destination-api-key"
	destinationApiSecretFlagName       = "destination-api-secret"
	destinationBootstrapServerFlagName = "destination-bootstrap-server"
	destinationClusterIdFlagName       = "destination-cluster-id"
	noValidateFlagName                 = "no-validate"
	sourceBootstrapServerFlagName      = "source-bootstrap-server"
	sourceClusterIdFlagName            = "source-cluster-id"
)

const (
	bootstrapServersPropertyName = "bootstrap.servers"
	saslJaasConfigPropertyName   = "sasl.jaas.config"
	saslMechanismPropertyName    = "sasl.mechanism"
	securityProtocolPropertyName = "security.protocol"
)

func (c *linkCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <link>",
		Short: "Create a new cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}

	example1 := examples.Example{Text: "Create a cluster link, using a configuration file."}
	example2 := examples.Example{Text: "Create a cluster link using command line flags."}
	if c.cfg.IsCloudLogin() {
		example1.Code = "confluent kafka link create my-link --source-cluster-id lkc-123456 --config-file config.txt"
		example2.Code = "confluent kafka link create my-link --source-cluster-id lkc-123456 --source-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret"
	} else {
		example1.Code = "confluent kafka link create my-link --destination-cluster-id 123456789 --config-file config.txt"
		example2.Code = "confluent kafka link create my-link --destination-cluster-id 123456789 --destination-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret"
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	// As of now, only CP --> CC links are supported.
	if c.cfg.IsCloudLogin() {
		cmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
		cmd.Flags().String(sourceBootstrapServerFlagName, "", "Bootstrap server address of the source cluster. Can alternatively be set in the config file using key bootstrap.servers.")
		cmd.Flags().String(destinationClusterIdFlagName, "", "Destination cluster ID for source initiated cluster links.")
		cmd.Flags().String(destinationBootstrapServerFlagName, "", "Bootstrap server address of the destination cluster for source initiated cluster links. Can alternatively be set in the config file using key bootstrap.servers.")
	} else {
		cmd.Flags().String(destinationClusterIdFlagName, "", "Destination cluster ID.")
		cmd.Flags().String(destinationBootstrapServerFlagName, "", "Bootstrap server address of the destination cluster. Can alternatively be set in the config file using key bootstrap.servers.")
	}

	cmd.Flags().String(sourceApiKeyFlagName, "", "An API key for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(sourceApiSecretFlagName, "", "An API secret for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(destinationApiKeyFlagName, "", "An API key for connecting to the destination cluster. "+
		"If specified, the source initiated cluster will use SASL_SSL/PLAIN as its mechanism for the destination cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(destinationApiSecretFlagName, "", "An API secret for connecting to the destination cluster. "+
		"If specified, the source initiated cluster will use SASL_SSL/PLAIN as its mechanism for the destination cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link configuration. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cmd.Flags().Bool(dryrunFlagName, false, "DEPRECATED: Validate a link, but do not create it (this flag is no longer active).")
	cmd.Flags().Bool(noValidateFlagName, false, "DEPRECATED: Create a link even if the source cluster cannot be reached (this flag is no longer active).")

	if c.cfg.IsCloudLogin() {
		pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	if !c.cfg.IsCloudLogin() {
		_ = cmd.MarkFlagRequired(destinationClusterIdFlagName)
	}

	return cmd
}

func (c *linkCommand) create(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	var sourceClusterId string
	var destinationClusterId string
	var bootstrapServer string
	var err error
	sourceClusterId, destinationClusterId, bootstrapServer, err = c.clusterIdsAndBootstrapServer(cmd)
	if err != nil {
		return err
	}

	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configMap := make(map[string]string)
	if configFile != "" {
		configMap, err = properties.FileToMap(configFile)
		if err != nil {
			return err
		}
	}

	apiKey, apiSecret, err := c.apiKeyAndSecret(cmd)
	if err != nil {
		return err
	}

	if bootstrapServer != "" {
		configMap[bootstrapServersPropertyName] = bootstrapServer
	}

	if apiKey != "" && apiSecret != "" {
		configMap[securityProtocolPropertyName] = "SASL_SSL"
		configMap[saslMechanismPropertyName] = "PLAIN"
		configMap[saslJaasConfigPropertyName] = fmt.Sprintf(`org.apache.kafka.common.security.plain.PlainLoginModule required username="%s" password="%s";`, apiKey, apiSecret)
	} else if apiKey != "" || apiSecret != "" {
		return errors.New("--source-api-key and --source-api-secret must be supplied together")
	}

	client, ctx, clusterId, err := c.getKafkaRestComponents(cmd)
	if err != nil {
		return err
	}

	data := kafkarestv3.CreateLinkRequestData{Configs: toCreateTopicConfigs(configMap)}
	if sourceClusterId != "" {
		data.SourceClusterId = sourceClusterId
	} else if destinationClusterId != "" {
		data.DestinationClusterId = destinationClusterId
	}

	opts := &kafkarestv3.CreateKafkaLinkOpts{CreateLinkRequestData: optional.NewInterface(data)}

	if httpResp, err := client.ClusterLinkingV3Api.CreateKafkaLink(ctx, clusterId, linkName, opts); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	utils.Printf(cmd, errors.CreatedLinkMsg, linkName)
	return nil
}

func (c *linkCommand) clusterIdsAndBootstrapServer(cmd *cobra.Command) (string, string, string, error) {
	var sourceClusterId string
	var destinationClusterId string
	var bootstrapServer string
	var err error
	if c.cfg.IsCloudLogin() {
		bootstrapServer, err = cmd.Flags().GetString(sourceBootstrapServerFlagName)
		if err != nil {
			return "", "", "", err
		}
		if bootstrapServer == "" {
			bootstrapServer, err = cmd.Flags().GetString(destinationBootstrapServerFlagName)
			if err != nil {
				return "", "", "", err
			}
		}
		sourceClusterId, err = cmd.Flags().GetString(sourceClusterIdFlagName)
		if err != nil {
			return "", "", "", err
		}
		if sourceClusterId == "" {
			destinationClusterId, err = cmd.Flags().GetString(destinationClusterIdFlagName)
			if err != nil {
				return "", "", "", err
			}
		}
	} else {
		bootstrapServer, err = cmd.Flags().GetString(destinationBootstrapServerFlagName)
		if err != nil {
			return "", "", "", err
		}
		destinationClusterId, err = cmd.Flags().GetString(destinationClusterIdFlagName)
		if err != nil {
			return "", "", "", err
		}
	}
	return sourceClusterId, destinationClusterId, bootstrapServer, nil
}

func (c *linkCommand) apiKeyAndSecret(cmd *cobra.Command) (string, string, error) {
	var err error

	var apiKey string
	apiKey, err = cmd.Flags().GetString(sourceApiKeyFlagName)
	if err != nil {
		return "", "", err
	}
	if apiKey == "" {
		apiKey, err = cmd.Flags().GetString(destinationApiKeyFlagName)
	}
	if err != nil {
		return "", "", err
	}

	var apiSecret string
	apiSecret, err = cmd.Flags().GetString(sourceApiSecretFlagName)
	if err != nil {
		return "", "", err
	}
	if apiSecret == "" {
		apiSecret, err = cmd.Flags().GetString(destinationApiSecretFlagName)
	}
	if err != nil {
		return "", "", err
	}
	return apiKey, apiSecret, nil
}
