package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newConfigurationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Kafka cluster configurations.",
		Args:  cobra.NoArgs,
		RunE:  c.configurationUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update Kafka cluster configuration "auto.create.topics.enable" to "true".`,
				Code: "confluent kafka cluster configuration update --config auto.create.topics.enable=true",
			},
		),
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides with form "key=value".`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *clusterCommand) configurationUpdate(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	configMap, err := properties.ConfigFlagToMap(config)
	if err != nil {
		return err
	}

	data := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(config))
	i := 0
	for key, value := range configMap {
		data[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: *kafkarestv3.NewNullableString(kafkarestv3.PtrString(value)),
		}
		i++
	}

	req := kafkarestv3.AlterConfigBatchRequestData{Data: data}
	if err := kafkaREST.CloudClient.UpdateKafkaClusterConfigs(cluster.ID, req); err != nil {
		return err
	}

	output.Println(formatUpdateOutput(configMap))

	return nil
}

func formatUpdateOutput(config map[string]string) string {
	names := types.GetSortedKeys(config)

	configuration := "configuration"
	if len(names) > 1 {
		configuration += "s"
	}

	return fmt.Sprintf("Successfully requested to update %s %s.", configuration, utils.ArrayToCommaDelimitedString(names, "and"))
}
