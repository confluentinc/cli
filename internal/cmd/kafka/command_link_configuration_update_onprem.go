package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
				Code: "confluent kafka link configuration update my-link --config-file my-config.txt",
			},
		),
	}

	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired(configFileFlagName))

	return cmd
}

func (c *linkCommand) configurationUpdateOnPrem(cmd *cobra.Command, args []string) error {
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

	client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(client, ctx)
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
