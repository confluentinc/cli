package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *linkCommand) describeOnPrem(cmd *cobra.Command, args []string) error {
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
		return kafkaRestError(pcmd.GetCPKafkaRestBaseUrl(client), err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, describeLinkConfigFields, humanDescribeLinkConfigFields, structuredDescribeLinkConfigFields)
	if err != nil {
		return err
	}

	outputWriter.AddElement(&LinkConfigWriter{
		ConfigName:  "dest.cluster.id",
		ConfigValue: listLinkConfigsRespData.Data[0].ClusterId,
		ReadOnly:    true,
		Sensitive:   true,
		Source:      "",
		Synonyms:    nil,
	})

	for _, config := range listLinkConfigsRespData.Data {
		outputWriter.AddElement(&LinkConfigWriter{
			ConfigName:  config.Name,
			ConfigValue: config.Value,
			ReadOnly:    config.ReadOnly,
			Sensitive:   config.Sensitive,
			Source:      config.Source,
			Synonyms:    config.Synonyms,
		})
	}
	return outputWriter.Out()
}
