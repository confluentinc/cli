package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *linkCommand) newConfigurationUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <link>",
		Short: "Update cluster link configurations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationUpdateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update configuration values for the cluster link "my-link".`,
				Code: "confluent kafka link configuration update my-link --config my-config.txt",
			},
		),
	}

	pcmd.AddConfigFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	// Deprecated
	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link configuration overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cobra.CheckErr(cmd.Flags().MarkHidden(configFileFlagName))
	cmd.MarkFlagsMutuallyExclusive("config", configFileFlagName)

	return cmd
}

func (c *linkCommand) configurationUpdateOnPrem(cmd *cobra.Command, args []string) error {
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

	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	opts := &kafkarestv3.UpdateKafkaLinkConfigBatchOpts{
		AlterConfigBatchRequestData: optional.NewInterface(toAlterConfigBatchRequestDataOnPrem(configMap)),
	}

	if httpResp, err := client.ClusterLinkingV3Api.UpdateKafkaLinkConfigBatch(ctx, clusterId, linkName, opts); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	output.Printf(errors.UpdatedResourceMsg, resource.ClusterLink, linkName)
	return nil
}
