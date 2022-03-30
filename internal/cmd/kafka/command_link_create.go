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
	apiKeyFlagName                     = "source-api-key"
	apiSecretFlagName                  = "source-api-secret"
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
	example2 := examples.Example{Text: "Create a cluster link, using the command line argument helpers."}
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
		cmd.Flags().String(sourceBootstrapServerFlagName, "", "Bootstrap server address of the source cluster. Can alternatively be set in the config file using key bootstrap.servers.")
		cmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
	} else {
		cmd.Flags().String(destinationBootstrapServerFlagName, "", "Bootstrap server address of the destination cluster. Can alternatively be set in the config file using key bootstrap.servers.")
		cmd.Flags().String(destinationClusterIdFlagName, "", "Destination cluster ID.")
	}

	cmd.Flags().String(apiKeyFlagName, "", "An API key for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(apiSecretFlagName, "", "An API secret for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file.")
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config. "+
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

	return cmd
}

func (c *linkCommand) create(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	var bootstrapServer string
	var err error
	if c.cfg.IsCloudLogin() {
		bootstrapServer, err = cmd.Flags().GetString(sourceBootstrapServerFlagName)
	} else {
		bootstrapServer, err = cmd.Flags().GetString(destinationBootstrapServerFlagName)
	}
	if err != nil {
		return err
	}

	var sourceClusterId string
	var destinationClusterId string
	if c.cfg.IsCloudLogin() {
		sourceClusterId, err = cmd.Flags().GetString(sourceClusterIdFlagName)
	} else {
		destinationClusterId, err = cmd.Flags().GetString(destinationClusterIdFlagName)
	}
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

	apiKey, err := cmd.Flags().GetString(apiKeyFlagName)
	if err != nil {
		return err
	}

	apiSecret, err := cmd.Flags().GetString(apiSecretFlagName)
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
	if c.cfg.IsCloudLogin() {
		data.SourceClusterId = sourceClusterId
	} else {
		data.DestinationClusterId = destinationClusterId
	}

	opts := &kafkarestv3.CreateKafkaLinkOpts{CreateLinkRequestData: optional.NewInterface(data)}

	if httpResp, err := client.ClusterLinkingV3Api.CreateKafkaLink(ctx, clusterId, linkName, opts); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	utils.Printf(cmd, errors.CreatedLinkMsg, linkName)
	return nil
}
