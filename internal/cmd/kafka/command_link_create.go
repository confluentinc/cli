package kafka

import (
	"fmt"

	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	apiKeyFlagName                     = "source-api-key"
	apiSecretFlagName                  = "source-api-secret"
	noValidateFlagName                 = "no-validate"
	saslJaasConfigPropertyName         = "sasl.jaas.config"
	saslMechanismPropertyName          = "sasl.mechanism"
	securityProtocolPropertyName       = "security.protocol"
	sourceBootstrapServersFlagName     = "source-bootstrap-server"
	sourceBootstrapServersPropertyName = "bootstrap.servers"
	sourceClusterIdFlagName            = "source-cluster-id"
)

func (c *linkCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <link>",
		Short: "Create a new cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using supplied source URL and properties.",
				Code: "confluent kafka link create my-link --source-cluster-id lkc-abcde --source-bootstrap-server my-host:1234 --config-file config.txt",
			},
			examples.Example{
				Code: "confluent kafka link create my-link --source-cluster-id lkc-abcde --source-bootstrap-server my-host:1234 --source-api-key my-key --source-api-secret my-secret",
			},
		),
	}

	cmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-server address of the source cluster.")
	cmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
	cmd.Flags().String(apiKeyFlagName, "", "An API key for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file. "+
		"Must be used with --source-api-secret.")
	cmd.Flags().String(apiSecretFlagName, "", "An API secret for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file. "+
		"Must be used with --source-api-key.")
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cmd.Flags().Bool(dryrunFlagName, false, "If set, will NOT actually create the link, but simply validates it.")
	cmd.Flags().Bool(noValidateFlagName, false, "If set, will create the link even if the source cluster cannot be reached with the supplied bootstrap server and credentials.")

	_ = cmd.MarkFlagRequired(sourceBootstrapServersFlagName)
	_ = cmd.MarkFlagRequired(sourceClusterIdFlagName)

	return cmd
}

func (c *linkCommand) create(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	bootstrapServers, err := cmd.Flags().GetString(sourceBootstrapServersFlagName)
	if err != nil {
		return err
	}

	sourceClusterId, err := cmd.Flags().GetString(sourceClusterIdFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	skipValidatingLink, err := cmd.Flags().GetBool(noValidateFlagName)
	if err != nil {
		return err
	}

	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configMap, err := utils.ReadConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	// Two optional flags: --source-api-key and --source-api-secret
	// 1. if I have neither flag set, then no change in behavior – use config-file as normal
	//
	// 2. if I have only 1 flag set, but not the other, then throw an error
	//
	// 3. if I have both set, then the CLI should add these configs on top of configs passed in config-file
	apiKey, err := cmd.Flags().GetString(apiKeyFlagName)
	if err != nil {
		return err
	}

	apiSecret, err := cmd.Flags().GetString(apiSecretFlagName)
	if err != nil {
		return err
	}

	// Overriding the security props by the flag value
	if apiKey != "" && apiSecret != "" {
		configMap[securityProtocolPropertyName] = "SASL_SSL"
		configMap[saslMechanismPropertyName] = "PLAIN"
		configMap[saslJaasConfigPropertyName] = fmt.Sprintf(
			"org.apache.kafka.common.security.plain.PlainLoginModule required "+
				"username=\"%s\" "+
				"password=\"%s\";", apiKey, apiSecret)
	} else if apiKey != "" {
		return errors.New("--source-api-key and --source-api-secret must be used together. " +
			"You cannot pass in one without the other.")
	}

	// Overriding the bootstrap server prop by the flag value
	configMap[sourceBootstrapServersPropertyName] = bootstrapServers

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	createLinkOpt := &kafkarestv3.ClustersClusterIdLinksPostOpts{
		ValidateOnly: optional.NewBool(validateOnly),
		ValidateLink: optional.NewBool(!skipValidatingLink),
		CreateLinkRequestData: optional.NewInterface(kafkarestv3.CreateLinkRequestData{
			SourceClusterId: sourceClusterId,
			Configs:         toCreateTopicConfigs(configMap),
		}),
	}

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksPost(
		kafkaREST.Context, lkc, linkName, createLinkOpt)

	if err == nil {
		msg := errors.CreatedLinkMsg
		if validateOnly {
			msg = errors.DryRunPrefix + msg
		}
		utils.Printf(cmd, msg, linkName)
		return nil
	}

	return handleOpenApiError(httpResp, err, kafkaREST)
}
