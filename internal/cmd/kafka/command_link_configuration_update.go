package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
				Code: "confluent kafka link configuration update my-link --config-file my-config.txt",
			},
		),
	}

	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link configuration overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired(configFileFlagName))

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
	if err != nil {
		return err
	}

	data := toAlterConfigBatchRequestData(configMap)

	if err := kafkaREST.CloudClient.UpdateKafkaLinkConfigBatch(linkName, data); err != nil {
		return err
	}

	output.Printf(errors.UpdatedResourceMsg, resource.ClusterLink, linkName)
	return nil
}
