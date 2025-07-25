package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *linkCommand) newConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <link>",
		Short:             "List cluster link configurations.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.configurationList,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) configurationList(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	configs, err := kafkaREST.CloudClient.ListKafkaLinkConfigs(args[0])
	if err != nil {
		return err
	}

	configList := append(configs, kafkarestv3.ListLinkConfigsResponseData{
		Name:        "dest.cluster.id",
		Value:       kafkaREST.GetClusterId(),
		IsReadOnly:  true,
		IsSensitive: true,
	})

	list := output.NewList(cmd)
	for _, config := range configList {
		list.Add(&linkConfigurationOut{
			ConfigName:  config.GetName(),
			ConfigValue: config.GetValue(),
			ReadOnly:    config.GetIsReadOnly(),
			Sensitive:   config.GetIsSensitive(),
			Source:      config.GetSource(),
			Synonyms:    config.GetSynonyms(),
		})
	}
	return list.Print()
}
