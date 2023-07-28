package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
			examples.Example{
				Text: `Create a mirror topic "src_my-topic" where "src_" is the prefix configured on the link:`,
				Code: "confluent kafka mirror create src_my-topic --link my-link --source-topic my-topic",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Int32(replicationFactorFlagName, 3, "Replication factor.")
	cmd.Flags().String(configFileFlagName, "", "Name of a file with additional topic configuration. Each property should be on its own line with the format: key=value.")
	cmd.Flags().String(sourceTopicFlagName, "", "Name of the source topic to be mirrored over the cluster link. Only required when there is a prefix configured on the link.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired(linkFlagName))

	return cmd
}

func (c *mirrorCommand) create(cmd *cobra.Command, args []string) error {
	mirrorTopicName := args[0]

	sourceTopicName, err := cmd.Flags().GetString(sourceTopicFlagName)
	if err != nil {
		return err
	}
	if sourceTopicName == "" {
		sourceTopicName = mirrorTopicName
	}

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

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	configs := toCreateTopicConfigs(configMap)
	createMirrorTopicRequestData := kafkarestv3.CreateMirrorTopicRequestData{
		SourceTopicName:   sourceTopicName,
		ReplicationFactor: &replicationFactor,
		Configs:           &configs,
	}
	// Only set the mirror topic if it differs from the source topic. This is for backwards compatibility: old versions
	// of ce-kafka-rest don't know about MirrorTopicName.
	if sourceTopicName != mirrorTopicName {
		createMirrorTopicRequestData.MirrorTopicName = &mirrorTopicName
	}

	if err := kafkaREST.CloudClient.CreateKafkaMirrorTopic(cluster.ID, linkName, createMirrorTopicRequestData); err != nil {
		return err
	}

	output.Printf(errors.CreatedResourceMsg, resource.MirrorTopic, mirrorTopicName)
	return nil
}
