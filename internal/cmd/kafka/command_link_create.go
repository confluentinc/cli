package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type linkMode int

const (
	Destination linkMode = iota
	Source
	Bidirectional
)

const (
	DESTINATION   = "DESTINATION"
	SOURCE        = "SOURCE"
	BIDIRECTIONAL = "BIDIRECTIONAL"
)

const (
	sourceApiKeyFlagName               = "source-api-key"
	sourceApiSecretFlagName            = "source-api-secret"
	destinationApiKeyFlagName          = "destination-api-key"
	destinationApiSecretFlagName       = "destination-api-secret"
	remoteApiKeyFlagName               = "remote-api-key"
	remoteApiSecretFlagName            = "remote-api-secret"
	localApiKeyFlagName                = "local-api-key"
	localApiSecretFlagName             = "local-api-secret"
	destinationBootstrapServerFlagName = "destination-bootstrap-server"
	destinationClusterIdFlagName       = "destination-cluster"
	noValidateFlagName                 = "no-validate"
	sourceBootstrapServerFlagName      = "source-bootstrap-server"
	sourceClusterIdFlagName            = "source-cluster"
	remoteBootstrapServerFlagName      = "remote-bootstrap-server"
	remoteClusterIdFlagName            = "remote-cluster"

	authHelperMsg = "If specified, the cluster will use SASL_SSL with PLAIN SASL as its mechanism for authentication. " +
		"If you wish to use another authentication mechanism, do not specify this flag, " +
		"and add the security configurations in the configuration file."
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
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using a configuration file.",
				Code: "confluent kafka link create my-link --source-cluster lkc-123456 --config-file config.txt",
			},
			examples.Example{
				Text: "Create a cluster link using command line flags.",
				Code: "confluent kafka link create my-link --source-cluster lkc-123456 --source-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret",
			},
		),
	}

	cmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
	cmd.Flags().String(sourceBootstrapServerFlagName, "", `Bootstrap server address of the source cluster. Can alternatively be set in the configuration file using key "bootstrap.servers".`)
	cmd.Flags().String(destinationClusterIdFlagName, "", "Destination cluster ID for source initiated cluster links.")
	cmd.Flags().String(destinationBootstrapServerFlagName, "", `Bootstrap server address of the destination cluster for source initiated cluster links. Can alternatively be set in the configuration file using key "bootstrap.servers".`)
	cmd.Flags().String(remoteClusterIdFlagName, "", "Remote cluster ID for bidirectional cluster links.")
	cmd.Flags().String(remoteBootstrapServerFlagName, "", `Bootstrap server address of the remote cluster for bidirectional links. Can alternatively be set in the configuration file using key "bootstrap.servers".`)
	cmd.Flags().String(sourceApiKeyFlagName, "", "An API key for the source cluster. For links at destination cluster this is used for remote cluster authentication. For links at source cluster this is used for local cluster authentication. "+authHelperMsg)
	cmd.Flags().String(sourceApiSecretFlagName, "", "An API secret for the source cluster. For links at destination cluster this is used for remote cluster authentication. For links at source cluster this is used for local cluster authentication. "+authHelperMsg)
	cmd.Flags().String(destinationApiKeyFlagName, "", "An API key for the destination cluster. This is used for remote cluster authentication links at the source cluster. "+authHelperMsg)
	cmd.Flags().String(destinationApiSecretFlagName, "", "An API secret for the destination cluster. This is used for remote cluster authentication for links at the source cluster. "+authHelperMsg)
	cmd.Flags().String(remoteApiKeyFlagName, "", "An API key for the remote cluster for bidirectional links. This is used for remote cluster authentication. "+authHelperMsg)
	cmd.Flags().String(remoteApiSecretFlagName, "", "An API secret for the remote cluster for bidirectional links. This is used for remote cluster authentication. "+authHelperMsg)
	cmd.Flags().String(localApiKeyFlagName, "", "An API key for the local cluster for bidirectional links. This is used for local cluster authentication if remote link's connection mode is Inbound. "+authHelperMsg)
	cmd.Flags().String(localApiSecretFlagName, "", "An API secret for the local cluster for bidirectional links. This is used for local cluster authentication if remote link's connection mode is Inbound. "+authHelperMsg)
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link configuration. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cmd.Flags().Bool(dryrunFlagName, false, "Validate a link, but do not create it.")
	cmd.Flags().Bool(noValidateFlagName, false, "Create a link even if the source cluster cannot be reached.")
	cmd.MarkFlagsRequiredTogether(sourceApiKeyFlagName, sourceApiSecretFlagName)
	cmd.MarkFlagsRequiredTogether(destinationApiKeyFlagName, destinationApiSecretFlagName)
	cmd.MarkFlagsRequiredTogether(localApiKeyFlagName, localApiSecretFlagName)
	cmd.MarkFlagsRequiredTogether(remoteApiKeyFlagName, remoteApiSecretFlagName)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *linkCommand) create(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	noValidate, err := cmd.Flags().GetBool(noValidateFlagName)
	if err != nil {
		return err
	}

	configMap, linkModeMetadata, err := c.getConfigMapAndLinkMode(configFile)
	if err != nil {
		return err
	}

	if err := c.addSecurityConfigToMap(cmd, linkModeMetadata, configMap); err != nil {
		return err
	}

	remoteClusterId, bootstrapServer, err := c.getRemoteClusterMetadata(cmd, linkModeMetadata)
	if err != nil {
		return err
	}

	if bootstrapServer != "" {
		configMap[bootstrapServersPropertyName] = bootstrapServer
	}

	configs := toCreateTopicConfigs(configMap)
	data := kafkarestv3.CreateLinkRequestData{Configs: &configs}

	if remoteClusterId != "" {
		switch linkModeMetadata.mode {
		case Destination:
			data.SourceClusterId = &remoteClusterId
		case Source:
			data.DestinationClusterId = &remoteClusterId
		case Bidirectional:
			data.RemoteClusterId = &remoteClusterId
		default:
			return unrecognizedLinkModeErr(linkModeMetadata.name)
		}
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if httpResp, err := kafkaREST.CloudClient.CreateKafkaLink(cluster.ID, linkName, !noValidate, dryRun, data); err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	msg := fmt.Sprintf(errors.CreatedLinkResourceMsg, resource.ClusterLink, linkName, linkConfigsCommandOutput(configMap))
	if dryRun {
		msg = utils.AddDryRunPrefix(msg)
	}
	output.Print(msg)

	return nil
}

// An allow-list of configs we can print. We don't want to print/log sensitive configs.
var printableLinkConfigs = types.NewSet(
	"bootstrap.servers",
	"connection.mode",
	"link.mode",
	"consumer.offset.sync.enable",
	"consumer.offset.group.filters",
	"consumer.offset.sync.ms",
	"consumer.group.prefix.enable",
	"acl.sync.enable",
	"acl.filters",
	"acl.sync.ms",
	"auto.create.mirror.topics.enable",
	"auto.create.mirror.topics.filters",
	"cluster.link.prefix",
	"topic.config.sync.ms",
	"topic.config.sync.include",
	"mirror.start.offset.spec",
	"security.protocol",
	"sasl.mechanism",
	"local.security.protocol",
	"local.sasl.mechanism",
)

func linkConfigsCommandOutput(linkConfig map[string]string) string {
	filtered := make(map[string]string)
	for key, val := range linkConfig {
		if printableLinkConfigs.Contains(key) {
			filtered[key] = val
		}
	}
	return properties.CreateKeyValuePairs(filtered)
}

func getJaasValue(apiKey, apiSecret string) string {
	return fmt.Sprintf(`%s username="%s" password="%s";`, jaasConfigPrefix, apiKey, apiSecret)
}

func (c *linkCommand) getConfigMapAndLinkMode(configFile string) (map[string]string, *linkModeMetadata, error) {
	if configFile != "" {
		var linkMode linkMode
		configMap, err := properties.FileToMap(configFile)
		if err != nil {
			return nil, nil, err
		}
		linkModeStr, ok := configMap["link.mode"]
		if !ok {
			// Default is destination if no config value is provided.
			linkMode = Destination
			linkModeStr = DESTINATION
		} else {
			switch linkModeStr {
			case DESTINATION:
				linkMode = Destination
			case SOURCE:
				linkMode = Source
			case BIDIRECTIONAL:
				linkMode = Bidirectional
			default:
				return nil, &linkModeMetadata{linkMode, linkModeStr}, unrecognizedLinkModeErr(linkModeStr)
			}
		}
		configMap["link.mode"] = linkModeStr
		return configMap, &linkModeMetadata{linkMode, linkModeStr}, nil
	} else {
		configMap := make(map[string]string)
		configMap["link.mode"] = DESTINATION
		// Default is destination if no config file is provided.
		return configMap, &linkModeMetadata{Destination, DESTINATION}, nil
	}
}

func unrecognizedLinkModeErr(linkModeStr string) error {
	return errors.Errorf(`unrecognized link.mode "%s". Use %s, %s, or %s.`, linkModeStr, DESTINATION, SOURCE, BIDIRECTIONAL)
}

func (c *linkCommand) addSecurityConfigToMap(cmd *cobra.Command, linkModeMetadata *linkModeMetadata, configMap map[string]string) error {
	switch linkModeMetadata.mode {
	case Destination:
		return c.addDestInitiatedLinkSecurityConfigToMap(cmd, configMap)
	case Source:
		return c.addSourceInitiatedLinkSecurityConfigToMap(cmd, configMap)
	case Bidirectional:
		return c.addBidirectionalLinkSecurityConfigToMap(cmd, configMap)
	default:
		return unrecognizedLinkModeErr(linkModeMetadata.name)
	}
}

func (c *linkCommand) addBidirectionalLinkSecurityConfigToMap(cmd *cobra.Command, configMap map[string]string) error {
	remoteApiKey, err := cmd.Flags().GetString(remoteApiKeyFlagName)
	if err != nil {
		return err
	}
	remoteApiSecret, err := cmd.Flags().GetString(remoteApiSecretFlagName)
	if err != nil {
		return err
	}
	if remoteApiKey != "" && remoteApiSecret != "" {
		configMap[securityProtocolPropertyName] = saslSsl
		configMap[saslMechanismPropertyName] = plain
		configMap[saslJaasConfigPropertyName] = getJaasValue(remoteApiKey, remoteApiSecret)
	}
	localApiKey, err := cmd.Flags().GetString(localApiKeyFlagName)
	if err != nil {
		return err
	}
	localApiSecret, err := cmd.Flags().GetString(localApiSecretFlagName)
	if err != nil {
		return err
	}
	if localApiKey != "" && localApiSecret != "" {
		configMap[localListenerPropertyName] = saslSsl
		configMap[localSecurityProtocolPropertyName] = saslSsl
		configMap[localSaslMechanismPropertyName] = plain
		configMap[localSaslJaasConfigPropertyName] = getJaasValue(localApiKey, localApiSecret)
	}
	return nil
}

func (c *linkCommand) addDestInitiatedLinkSecurityConfigToMap(cmd *cobra.Command, configMap map[string]string) error {
	sourceApiKey, err := cmd.Flags().GetString(sourceApiKeyFlagName)
	if err != nil {
		return err
	}
	sourceApiSecret, err := cmd.Flags().GetString(sourceApiSecretFlagName)
	if err != nil {
		return err
	}
	if sourceApiKey != "" && sourceApiSecret != "" {
		configMap[securityProtocolPropertyName] = saslSsl
		configMap[saslMechanismPropertyName] = plain
		configMap[saslJaasConfigPropertyName] = getJaasValue(sourceApiKey, sourceApiSecret)
	} else if sourceApiKey != "" || sourceApiSecret != "" {
		return errors.New("--source-api-key and --source-api-secret must be supplied together")
	}
	return nil
}

func (c *linkCommand) addSourceInitiatedLinkSecurityConfigToMap(cmd *cobra.Command, configMap map[string]string) error {
	sourceApiKey, err := cmd.Flags().GetString(sourceApiKeyFlagName)
	if err != nil {
		return err
	}
	sourceApiSecret, err := cmd.Flags().GetString(sourceApiSecretFlagName)
	if err != nil {
		return err
	}
	if sourceApiKey != "" && sourceApiSecret != "" {
		configMap[localListenerPropertyName] = saslSsl
		configMap[localSecurityProtocolPropertyName] = saslSsl
		configMap[localSaslMechanismPropertyName] = plain
		configMap[localSaslJaasConfigPropertyName] = getJaasValue(sourceApiKey, sourceApiSecret)
	} else if sourceApiKey != "" || sourceApiSecret != "" {
		return errors.New("--source-api-key and --source-api-secret must be supplied together")
	}
	destinationApiKey, err := cmd.Flags().GetString(destinationApiKeyFlagName)
	if err != nil {
		return err
	}
	destinationApiSecret, err := cmd.Flags().GetString(destinationApiSecretFlagName)
	if err != nil {
		return err
	}
	if destinationApiKey != "" && destinationApiSecret != "" {
		// For source initiated links at source cluster, the credentials are for the remote cluster.
		configMap[securityProtocolPropertyName] = saslSsl
		configMap[saslMechanismPropertyName] = plain
		configMap[saslJaasConfigPropertyName] = getJaasValue(destinationApiKey, destinationApiSecret)
	} else if destinationApiKey != "" || destinationApiSecret != "" {
		return errors.New("--destination-api-key and --destination-api-secret must be supplied together")
	}
	return nil
}

func (c *linkCommand) getRemoteClusterMetadata(cmd *cobra.Command, linkModeMetadata *linkModeMetadata) (string, string, error) {
	switch linkModeMetadata.mode {
	case Destination:
		// For links at destination cluster, look for the source bootstrap servers and cluster ID.
		bootstrapServer, err := cmd.Flags().GetString(sourceBootstrapServerFlagName)
		if err != nil {
			return "", "", err
		}
		remoteClusterId, err := cmd.Flags().GetString(sourceClusterIdFlagName)
		if err != nil {
			return "", "", err
		}
		return remoteClusterId, bootstrapServer, nil
	case Source:
		// For links at source cluster, look for the destination bootstrap servers and cluster ID.
		bootstrapServer, err := cmd.Flags().GetString(destinationBootstrapServerFlagName)
		if err != nil {
			return "", "", err
		}
		remoteClusterId, err := cmd.Flags().GetString(destinationClusterIdFlagName)
		if err != nil {
			return "", "", err
		}
		return remoteClusterId, bootstrapServer, nil
	case Bidirectional:
		bootstrapServer, err := cmd.Flags().GetString(remoteBootstrapServerFlagName)
		if err != nil {
			return "", "", err
		}
		remoteClusterId, err := cmd.Flags().GetString(remoteClusterIdFlagName)
		if err != nil {
			return "", "", err
		}
		return remoteClusterId, bootstrapServer, nil
	default:
		return "", "", unrecognizedLinkModeErr(linkModeMetadata.name)
	}
}

type linkModeMetadata struct {
	mode linkMode
	name string
}
