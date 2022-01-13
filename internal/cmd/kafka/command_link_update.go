package kafka

import (
	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *linkCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <link>",
		Short: "Update link configs.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for the cluster link `my-link`.",
				Code: "confluent kafka link update my-link --config-file my-config.txt",
			},
		),
	}

	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")

	if c.cfg.IsOnPremLogin() {
		cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired(configFileFlagName)

	return cmd
}

func (c *linkCommand) update(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configsMap, err := utils.ReadConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	if len(configsMap) == 0 {
		return errors.New(errors.EmptyConfigErrorMsg)
	}

	client, ctx, clusterId, err := c.getKafkaRestComponents(cmd)
	if err != nil {
		return err
	}

	opts := &kafkarestv3.ClustersClusterIdLinksLinkNameConfigsalterPutOpts{
		AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{
			Data: toAlterConfigBatchRequestData(configsMap),
		}),
	}

	if httpResp, err := client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameConfigsalterPut(ctx, clusterId, linkName, opts); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	utils.Printf(cmd, errors.UpdatedLinkMsg, linkName)
	return nil
}
