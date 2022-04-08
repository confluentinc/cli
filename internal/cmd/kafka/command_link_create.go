package kafka

import (
	"fmt"
	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type linkMode int

const (
	Destination linkMode = iota
	Source
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
	bootstrapServersPropertyName      = "bootstrap.servers"
	saslJaasConfigPropertyName        = "sasl.jaas.config"
	saslMechanismPropertyName         = "sasl.mechanism"
	securityProtocolPropertyName      = "security.protocol"
	localListenerPropertyName         = "local.listener"
	localSecurityProtocolPropertyName = "local.security.protocol"
	localSaslMechanismPropertyName    = "local.sasl.mechanism"
	localSaslJaasConfigPropertyName   = "local.sasl.jaas.config"
)

const (
	saslSsl          = "SASL_SSL"
	plain            = "PLAIN"
	jaasConfigPrefix = "org.apache.kafka.common.security.plain.PlainLoginModule required"
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
		cmd.Flags().String(destinationBootstrapServerFlagName, "", `Bootstrap server address of the destination cluster for source initiated cluster links. Can alternatively be set in the config file using key "bootstrap.servers".`)
	} else {
		cmd.Flags().String(destinationClusterIdFlagName, "", "Destination cluster ID.")
		cmd.Flags().String(destinationBootstrapServerFlagName, "", "Bootstrap server address of the destination cluster. Can alternatively be set in the config file using key bootstrap.servers.")
	}

	cmd.Flags().String(sourceApiKeyFlagName, "", "An API key for the source cluster. "+
		"For destination initiated links this is used for remote cluster authentication. For source initiated links this is used for local cluster authentication. "+
		"If specified, the cluster will use SASL_SSL/PLAIN as its mechanism for authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(sourceApiSecretFlagName, "", "An API secret for the source cluster. "+
		"For destination initiated links this is used for remote cluster authentication. For source initiated links this is used for local cluster authentication. "+
		"If specified, the cluster will use SASL_SSL/PLAIN as its mechanism for authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(destinationApiKeyFlagName, "", "An API key for the destination cluster. "+
		"This is used for remote cluster authentication for source initiated links. "+
		"If specified, cluster will use SASL_SSL/PLAIN as its mechanism for authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(destinationApiSecretFlagName, "", "An API secret for the destination cluster. "+
		"This is used for remote cluster authentication for source initiated links. "+
		"If specified, the cluster will use SASL_SSL/PLAIN as its mechanism for authentication. "+
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

	var err error
	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configMapAndLinkMode, err := c.getConfigMapAndLinkMode(configFile)
	if err != nil {
		return err
	}
	configMap := configMapAndLinkMode.configMap
	linkMode := configMapAndLinkMode.linkMode

	if !c.cfg.IsCloudLogin() {
		// On prem we only support source initiated links currently.
		if linkMode != Source {
			return errors.New("Source initiated links are only supported type for on-prem")
		}
	}

	err = c.addSecurityConfigToMap(cmd, linkMode, configMap)
	if err != nil {
		return err
	}

	remoteClusterMetadata, err := c.getRemoteClusterMetadata(cmd, linkMode)
	if err != nil {
		return err
	}

	if remoteClusterMetadata.bootstrapServer != "" {
		configMap[bootstrapServersPropertyName] = remoteClusterMetadata.bootstrapServer
	}

	data := kafkarestv3.CreateLinkRequestData{Configs: toCreateTopicConfigs(configMap)}
	if linkMode == Destination && remoteClusterMetadata.remoteClusterId != "" {
		data.SourceClusterId = remoteClusterMetadata.remoteClusterId
	} else if remoteClusterMetadata.remoteClusterId != "" {
		data.DestinationClusterId = remoteClusterMetadata.remoteClusterId
	}

	opts := &kafkarestv3.CreateKafkaLinkOpts{CreateLinkRequestData: optional.NewInterface(data)}

	client, ctx, clusterId, err := c.getKafkaRestComponents(cmd)
	if err != nil {
		return err
	}

	if httpResp, err := client.ClusterLinkingV3Api.CreateKafkaLink(ctx, clusterId, linkName, opts); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	utils.Printf(cmd, errors.CreatedLinkMsg, linkName)
	return nil
}

func getJaasValue(apiKey, apiSecret string) string {
	return fmt.Sprintf(jaasConfigPrefix+` username="%s" password="%s";`, apiKey, apiSecret)
}

func (c *linkCommand) getConfigMapAndLinkMode(configFile string) (*configMapAndLinkMode, error) {
	configMap := make(map[string]string)
	var err error
	var linkMode linkMode
	if configFile != "" {
		configMap, err = properties.FileToMap(configFile)
		if err != nil {
			return nil, err
		}
		linkModeStr, ok := configMap["link.mode"]
		if !ok {
			// Default is destination if no config value is provided.
			linkMode = Destination
		} else if strings.EqualFold(linkModeStr, "DESTINATION") {
			linkMode = Destination
		} else if strings.EqualFold(linkModeStr, "SOURCE") {
			linkMode = Source
		} else {
			return nil, errors.Errorf(`unrecognized link.mode "%s". Use DESTINATION or SOURCE.`, linkModeStr)
		}
	} else {
		// Default is destination if no config file is provided.
		linkMode = Destination
	}
	return &configMapAndLinkMode{configMap, linkMode}, nil
}

type configMapAndLinkMode struct {
	configMap map[string]string
	linkMode  linkMode
}

func (c *linkCommand) addSecurityConfigToMap(cmd *cobra.Command, linkMode linkMode, configMap map[string]string) error {
	sourceApiKey, err := cmd.Flags().GetString(sourceApiKeyFlagName)
	if err != nil {
		return err
	}
	sourceApiSecret, err := cmd.Flags().GetString(sourceApiSecretFlagName)
	if err != nil {
		return err
	}
	if sourceApiKey != "" && sourceApiSecret != "" {
		if linkMode == Destination {
			// For destination initiated links, the credentials are for the remote cluster.
			configMap[securityProtocolPropertyName] = saslSsl
			configMap[saslMechanismPropertyName] = plain
			configMap[saslJaasConfigPropertyName] = getJaasValue(sourceApiKey, sourceApiSecret)
		} else {
			// For source initiated links, the credentials are for the local cluster.
			configMap[localListenerPropertyName] = saslSsl
			configMap[localSecurityProtocolPropertyName] = saslSsl
			configMap[localSaslMechanismPropertyName] = plain
			configMap[localSaslJaasConfigPropertyName] = getJaasValue(sourceApiKey, sourceApiSecret)
		}
	} else if sourceApiKey != "" || sourceApiSecret != "" {
		return errors.New("--source-api-key and --source-api-secret must be supplied together")
	}

	if linkMode == Source {
		destinationApiKey, err := cmd.Flags().GetString(destinationApiKeyFlagName)
		if err != nil {
			return err
		}
		destinationApiSecret, err := cmd.Flags().GetString(destinationApiSecretFlagName)
		if err != nil {
			return err
		}
		if destinationApiKey != "" && destinationApiSecret != "" {
			// For source initiated links, the credentials are for the remote cluster.
			configMap[securityProtocolPropertyName] = saslSsl
			configMap[saslMechanismPropertyName] = plain
			configMap[saslJaasConfigPropertyName] = getJaasValue(destinationApiKey, destinationApiSecret)
		} else if destinationApiKey != "" || destinationApiSecret != "" {
			return errors.New("--destination-api-key and --destination-api-secret must be supplied together")
		}
	}
	return nil
}

type remoteClusterMetadata struct {
	remoteClusterId string
	bootstrapServer string
}

func (c *linkCommand) getRemoteClusterMetadata(cmd *cobra.Command, linkMode linkMode) (*remoteClusterMetadata, error) {
	if linkMode == Destination {
		// For destination initiated links, look for the source bootstrap servers and cluster ID.
		bootstrapServer, err := cmd.Flags().GetString(sourceBootstrapServerFlagName)
		if err != nil {
			return nil, err
		}
		remoteClusterId, err := cmd.Flags().GetString(sourceClusterIdFlagName)
		if err != nil {
			return nil, err
		}
		return &remoteClusterMetadata{remoteClusterId, bootstrapServer}, nil
	} else {
		// For source initiated links, look for the destination bootstrap servers and cluster ID.
		bootstrapServer, err := cmd.Flags().GetString(destinationBootstrapServerFlagName)
		if err != nil {
			return nil, err
		}
		remoteClusterId, err := cmd.Flags().GetString(destinationClusterIdFlagName)
		if err != nil {
			return nil, err
		}
		return &remoteClusterMetadata{remoteClusterId, bootstrapServer}, nil
	}
}
