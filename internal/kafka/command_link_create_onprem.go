package kafka

import (
	"fmt"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *linkCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <link>",
		Short: "Create a new cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using a configuration file.",
				Code: "confluent kafka link create my-link --destination-cluster 123456789 --config config.txt",
			},
			examples.Example{
				Text: "Create a source cluster link using command line flags.",
				Code: "confluent kafka link create my-link --destination-cluster 123456789 --destination-bootstrap-server my-host:1234 --destination-api-key remote-key --destination-api-secret remote-secret --source-api-key local-key --source-api-secret local-secret --config link.mode=SOURCE,connection.mode=OUTBOUND",
			},
			examples.Example{
				Text: "Create a bidirectional cluster link using command line flags.",
				Code: "confluent kafka link create my-link --remote-cluster 123456789 --remote-bootstrap-server my-host:1234 --remote-api-key remote-key --remote-api-secret remote-secret --local-api-key local-key --local-api-secret local-secret --config link.mode=BIDIRECTIONAL",
			},
		),
	}

	cmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
	cmd.Flags().String(sourceBootstrapServerFlagName, "", `Bootstrap server address of the source cluster. Can alternatively be set in the configuration file using key "bootstrap.servers".`)
	cmd.Flags().String(destinationClusterIdFlagName, "", "Destination cluster ID.")
	cmd.Flags().String(destinationBootstrapServerFlagName, "", `Bootstrap server address of the destination cluster. Can alternatively be set in the configuration file using key "bootstrap.servers".`)
	cmd.Flags().String(remoteClusterIdFlagName, "", "Remote cluster ID for bidirectional cluster links.")
	cmd.Flags().String(remoteBootstrapServerFlagName, "", `Bootstrap server address of the remote cluster for bidirectional links. Can alternatively be set in the configuration file using key "bootstrap.servers".`)
	cmd.Flags().String(sourceApiKeyFlagName, "", "An API key for the source cluster. For links at destination cluster, this is used for remote cluster authentication. For links at source cluster, this is used for local cluster authentication. "+authHelperMsg)
	cmd.Flags().String(sourceApiSecretFlagName, "", "An API secret for the source cluster. For links at destination cluster, this is used for remote cluster authentication. For links at source cluster, this is used for local cluster authentication. "+authHelperMsg)
	cmd.Flags().String(destinationApiKeyFlagName, "", "An API key for the destination cluster. This is used for remote cluster authentication links at the source cluster. "+authHelperMsg)
	cmd.Flags().String(destinationApiSecretFlagName, "", "An API secret for the destination cluster. This is used for remote cluster authentication for links at the source cluster. "+authHelperMsg)
	cmd.Flags().String(remoteApiKeyFlagName, "", "An API key for the remote cluster for bidirectional links. This is used for remote cluster authentication. "+authHelperMsg)
	cmd.Flags().String(remoteApiSecretFlagName, "", "An API secret for the remote cluster for bidirectional links. This is used for remote cluster authentication. "+authHelperMsg)
	cmd.Flags().String(localApiKeyFlagName, "", "An API key for the local cluster for bidirectional links. This is used for local cluster authentication if remote link's connection mode is Inbound. "+authHelperMsg)
	cmd.Flags().String(localApiSecretFlagName, "", "An API secret for the local cluster for bidirectional links. This is used for local cluster authentication if remote link's connection mode is Inbound. "+authHelperMsg)
	pcmd.AddConfigFlag(cmd)
	cmd.Flags().Bool(dryrunFlagName, false, "Validate a link, but do not create it.")
	cmd.Flags().Bool(noValidateFlagName, false, "Create a link even if the source cluster cannot be reached.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	// Deprecated
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link configuration. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cobra.CheckErr(cmd.Flags().MarkHidden(configFileFlagName))
	cmd.MarkFlagsMutuallyExclusive("config", configFileFlagName)

	// Hide flags for destination links; removing them causes a flag read error instead of the proper error
	cobra.CheckErr(cmd.Flags().MarkHidden(sourceClusterIdFlagName))
	cobra.CheckErr(cmd.Flags().MarkHidden(sourceBootstrapServerFlagName))

	cmd.MarkFlagsOneRequired(sourceClusterIdFlagName, destinationClusterIdFlagName, remoteClusterIdFlagName)

	return cmd
}

func (c *linkCommand) createOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	// Deprecated
	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}
	if configFile != "" {
		config = []string{configFile}
	}

	configMap, err := properties.GetMap(config)
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

	configMap, linkModeMetadata, err := c.getConfigMapAndLinkMode(configMap)
	if err != nil {
		return err
	}

	linkMode := linkModeMetadata.mode
	if linkMode != Source && linkMode != Bidirectional {
		return fmt.Errorf("only source-initiated or bidirectional links can be created for Confluent Platform from the CLI")
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

	data := kafkarestv3.CreateLinkRequestData{Configs: toCreateTopicConfigsOnPrem(configMap)}
	if remoteClusterId != "" {
		switch linkModeMetadata.mode {
		case Destination:
			data.SourceClusterId = remoteClusterId
		case Source:
			data.DestinationClusterId = remoteClusterId
		case Bidirectional:
			data.RemoteClusterId = remoteClusterId
		default:
			return unrecognizedLinkModeErr(linkModeMetadata.name)
		}
	}

	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	opts := &kafkarestv3.CreateKafkaLinkOpts{
		ValidateOnly:          optional.NewBool(dryRun),
		ValidateLink:          optional.NewBool(!noValidate),
		CreateLinkRequestData: optional.NewInterface(data),
	}

	if httpResp, err := client.ClusterLinkingV3Api.CreateKafkaLink(ctx, clusterId, linkName, opts); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	msg := fmt.Sprintf(createdLinkResourceMsg, resource.ClusterLink, linkName)
	if dryRun {
		msg = utils.AddDryRunPrefix(msg)
	}

	output.Println(c.Config.EnableColor, msg)
	output.Println(c.Config.EnableColor, linkConfigsCommandOutput(configMap))

	return nil
}

func getListFieldsOnPrem(includeTopics bool) []string {
	x := []string{"Name", "Id"}

	if includeTopics {
		x = append(x, "TopicName")
	}

	return append(x, "DestinationCluster", "RemoteCluster", "State", "Error", "ErrorMessage")
}
