package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *linkCommand) newConfigurationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <link>",
		Short: "Update cluster link configurations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update configuration values for the cluster link "my-link".`,
				Code: "confluent kafka link configuration update my-link --config-file my-config.txt",
			},
		),
	}

	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	_ = cmd.MarkFlagRequired(configFileFlagName)

	return cmd
}

func (c *linkCommand) configurationUpdate(cmd *cobra.Command, args []string) error {
	linkName := args[0]

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

	if len(configMap) == 0 {
		return errors.New(errors.EmptyConfigErrorMsg)
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	clusterId, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	data := toAlterConfigBatchRequestData(configMap)

	if httpResp, err := kafkaREST.CloudClient.UpdateKafkaLinkConfigBatch(clusterId, linkName, data); err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	output.Printf(errors.UpdatedResourceMsg, resource.ClusterLink, linkName)
	return nil
}
