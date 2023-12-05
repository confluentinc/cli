package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *linkCommand) newConfigurationListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <link>",
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
	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	configs, httpResp, err := client.ClusterLinkingV3Api.ListKafkaLinkConfigs(ctx, clusterId, args[0])
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	configList := configs.Data

	if len(configList) > 0 {
		configList = append(configList, kafkarestv3.ListLinkConfigsResponseData{
			Name:      "dest.cluster.id",
			Value:     configList[0].ClusterId,
			ReadOnly:  true,
			Sensitive: true,
		})
	}

	list := output.NewList(cmd)
	for _, config := range configList {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&linkConfigurationHumanOut{
				ConfigName:  config.Name,
				ConfigValue: config.Value,
				ReadOnly:    config.ReadOnly,
				Sensitive:   config.Sensitive,
				Source:      config.Source,
				Synonyms:    strings.Join(config.Synonyms, ", "),
			})
		} else {
			list.Add(&linkConfigurationSerializedOut{
				ConfigName:  config.Name,
				ConfigValue: config.Value,
				ReadOnly:    config.ReadOnly,
				Sensitive:   config.Sensitive,
				Source:      config.Source,
				Synonyms:    config.Synonyms,
			})
		}
	}
	return list.Print()
}
