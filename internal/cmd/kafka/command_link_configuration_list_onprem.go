package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *linkCommand) newConfigurationListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <name>",
		Short: "List cluster link configurations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationListOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) configurationListOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(client, ctx)
	if err != nil {
		return err
	}

	listLinkConfigsRespData, httpResp, err := client.ClusterLinkingV3Api.ListKafkaLinkConfigs(ctx, clusterId, linkName)
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	list := output.NewList(cmd)
	if len(listLinkConfigsRespData.Data) == 0 {
		return list.Print()
	}

	list.Add(&linkConfigurationOut{
		ConfigName:  "dest.cluster.id",
		ConfigValue: listLinkConfigsRespData.Data[0].ClusterId,
		ReadOnly:    true,
		Sensitive:   true,
	})

	for _, config := range listLinkConfigsRespData.Data {
		list.Add(&linkConfigurationOut{
			ConfigName:  config.Name,
			ConfigValue: config.Value,
			ReadOnly:    config.ReadOnly,
			Sensitive:   config.Sensitive,
			Source:      config.Source,
			Synonyms:    config.Synonyms,
		})
	}
	return list.Print()
}
