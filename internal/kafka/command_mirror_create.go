package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
	"github.com/confluentinc/cli/v3/pkg/resource"
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
				Code: "confluent kafka mirror create my-topic --link my-link --replication-factor 5 --config my-config.txt",
			},
			examples.Example{
				Text: `Create a mirror topic "src_my-topic" where "src_" is the prefix configured on the link:`,
				Code: "confluent kafka mirror create src_my-topic --link my-link --source-topic my-topic",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Int32("replication-factor", 3, "Replication factor.")
	pcmd.AddConfigFlag(cmd)
	cmd.Flags().String("source-topic", "", "Name of the source topic to be mirrored over the cluster link. Only required when there is a prefix configured on the link.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	// Deprecated
	cmd.Flags().String(configFileFlagName, "", "Name of a file with additional topic configuration. Each property should be on its own line with the format: key=value.")
	cobra.CheckErr(cmd.Flags().MarkHidden(configFileFlagName))
	cmd.MarkFlagsMutuallyExclusive("config", configFileFlagName)

	cobra.CheckErr(cmd.MarkFlagRequired("link"))

	return cmd
}

func (c *mirrorCommand) create(cmd *cobra.Command, args []string) error {
	mirrorTopicName := args[0]

	sourceTopic, err := cmd.Flags().GetString("source-topic")
	if err != nil {
		return err
	}
	if sourceTopic == "" {
		sourceTopic = mirrorTopicName
	}

	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}

	replicationFactor, err := cmd.Flags().GetInt32("replication-factor")
	if err != nil {
		return err
	}

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

	configMap, err := properties.GetMapWithJavaPropertyParsing(config)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	configs := toCreateTopicConfigs(configMap)
	data := kafkarestv3.CreateMirrorTopicRequestData{
		SourceTopicName:   sourceTopic,
		ReplicationFactor: kafkarestv3.PtrInt32(replicationFactor),
		Configs:           &configs,
	}
	// Only set the mirror topic if it differs from the source topic. This is for backwards compatibility: old versions
	// of ce-kafka-rest don't know about MirrorTopicName.
	if sourceTopic != mirrorTopicName {
		data.MirrorTopicName = kafkarestv3.PtrString(mirrorTopicName)
	}

	if err := kafkaREST.CloudClient.CreateKafkaMirrorTopic(link, data); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.CreatedResourceMsg, resource.MirrorTopic, mirrorTopicName)
	return nil
}
