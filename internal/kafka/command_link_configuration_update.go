package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *linkCommand) newConfigurationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <link>",
		Short:             "Update cluster link configurations.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.configurationUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update configuration values for the cluster link "my-link".`,
				Code: "confluent kafka link configuration update my-link --config my-config.txt",
			},
		),
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddConfigFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	// Deprecated
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link configuration overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cobra.CheckErr(cmd.Flags().MarkHidden(configFileFlagName))
	cmd.MarkFlagsMutuallyExclusive("config", configFileFlagName)

	return cmd
}

func (c *linkCommand) configurationUpdate(cmd *cobra.Command, args []string) error {
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

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	data := kafkarestv3.AlterConfigBatchRequestData{Data: toAlterConfigBatchRequestData(configMap)}
	if err := kafkaREST.CloudClient.UpdateKafkaLinkConfigBatch(linkName, data); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.ClusterLink, linkName)
	return nil
}
