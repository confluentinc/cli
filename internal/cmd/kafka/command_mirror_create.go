package kafka

import (
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *mirrorCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <source-topic-name>",
		Short: "Create a mirror topic under the link.",
		Long:  "Create a mirror topic under the link. The destination topic name is required to be the same as the source topic name.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a mirror topic "my-topic" under cluster link "my-link":`,
				Code: "confluent kafka mirror create my-topic --link my-link",
			},
			examples.Example{
				Text: "Create a mirror topic with a custom replication factor and configuration file:",
				Code: "confluent kafka mirror create my-topic --link my-link --replication-factor 5 --config-file my-config.txt",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "The name of the cluster link to attach to the mirror topic.")
	cmd.Flags().Int32(replicationFactorFlagName, 3, "Replication factor.")
	cmd.Flags().String(configFileFlagName, "", "Name of a file with additional topic configuration. Each property should be on its own line with the format: key=value.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired(linkFlagName)

	return cmd
}

func (c *mirrorCommand) create(cmd *cobra.Command, args []string) error {
	sourceTopicName := args[0]

	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	replicationFactor, err := cmd.Flags().GetInt32(replicationFactorFlagName)
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

	kafkaREST, err := c.GetCloudKafkaREST()
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

	createMirrorOpt := cloudkafkarest.CreateMirrorTopicRequestData{
		SourceTopicName:   sourceTopicName,
		ReplicationFactor: &replicationFactor,
		Configs:           toCloudCreateTopicConfigs(configMap),
	}

	req := kafkaREST.Client.ClusterLinkingV3Api.CreateKafkaMirrorTopic(kafkaREST.Context, lkc, linkName)
	httpResp, err := req.CreateMirrorTopicRequestData(createMirrorOpt).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}
	utils.Printf(cmd, errors.CreatedMirrorMsg, sourceTopicName)
	return nil
}
