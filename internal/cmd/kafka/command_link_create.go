package kafka

import (
	"fmt"
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

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
	}

	example1 := examples.Example{Text: "Create a cluster link, using a configuration file."}
	example2 := examples.Example{Text: "Create a cluster link, using an API key and secret."}
	if c.cfg.IsCloudLogin() {
		cmd.RunE = c.create
		example1.Code = "confluent kafka link create my-link --source-cluster-id lkc-123456 --source-bootstrap-server my-host:1234 --config-file config.txt"
		example2.Code = "confluent kafka link create my-link --source-cluster-id lkc-123456 --source-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret"
	} else {
		cmd.RunE = c.createOnPrem
		example1.Code = "confluent kafka link create my-link --destination-cluster-id 123456789 --destination-bootstrap-server my-host:1234 --config-file config.txt"
		example2.Code = "confluent kafka link create my-link --destination-cluster-id 123456789 --destination-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret"
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	// As of now, only CP --> CC links are supported.
	if c.cfg.IsCloudLogin() {
		cmd.Flags().String(sourceBootstrapServerFlagName, "", "Bootstrap server address of the source cluster.")
		cmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
	} else {
		cmd.Flags().String(destinationBootstrapServerFlagName, "", "Bootstrap server address of the destination cluster.")
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
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. "+
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

	if c.cfg.IsCloudLogin() {
		_ = cmd.MarkFlagRequired(sourceBootstrapServerFlagName)
		_ = cmd.MarkFlagRequired(sourceClusterIdFlagName)
	} else {
		_ = cmd.MarkFlagRequired(destinationBootstrapServerFlagName)
		_ = cmd.MarkFlagRequired(destinationClusterIdFlagName)
	}

	return cmd
}

func (c *linkCommand) create(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	bootstrapServer, err := cmd.Flags().GetString(sourceBootstrapServerFlagName)
	if err != nil {
		return err
	}

	sourceClusterId, err := cmd.Flags().GetString(sourceClusterIdFlagName)
	if err != nil {
		return err
	}

	configMap, err := c.parseConfigMap(bootstrapServer)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetCloudKafkaREST()
	if err != nil {
		return err
	}
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	clusterId := kafkaClusterConfig.ID

	data := cloudkafkarest.CreateLinkRequestData{SourceClusterId: &sourceClusterId, Configs: toCloudCreateTopicConfigs(configMap)}

	req := kafkaREST.Client.ClusterLinkingV3Api.CreateKafkaLink(kafkaREST.Context, clusterId)
	httpResp, err := req.CreateLinkRequestData(data).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	utils.Printf(cmd, errors.CreatedLinkMsg, linkName)
	return nil
}

func (c *linkCommand) parseConfigMap(bootstrapServer string) (map[string]string, error) {
	configMap := make(map[string]string)

	configFile, err := c.Flags().GetString(configFileFlagName)
	if err != nil {
		return configMap, err
	}

	if configFile != "" {
		configMap, err = properties.FileToMap(configFile)
		if err != nil {
			return configMap, err
		}
	}

	apiKey, err := c.Flags().GetString(apiKeyFlagName)
	if err != nil {
		return configMap, err
	}

	apiSecret, err := c.Flags().GetString(apiSecretFlagName)
	if err != nil {
		return configMap, err
	}

	configMap[bootstrapServersPropertyName] = bootstrapServer

	if apiKey != "" && apiSecret != "" {
		configMap[securityProtocolPropertyName] = "SASL_SSL"
		configMap[saslMechanismPropertyName] = "PLAIN"
		configMap[saslJaasConfigPropertyName] = fmt.Sprintf(`org.apache.kafka.common.security.plain.PlainLoginModule required username="%s" password="%s";`, apiKey, apiSecret)
	} else if apiKey != "" || apiSecret != "" {
		return configMap, errors.New("--source-api-key and --source-api-secret must be supplied together")
	}
	return configMap, nil
}