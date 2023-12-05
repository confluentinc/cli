package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
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
		Name:      "dest.cluster.id",
		Value:     kafkaREST.GetClusterId(),
		ReadOnly:  true,
		Sensitive: true,
	})

	list := output.NewList(cmd)
	for _, config := range configList {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&linkConfigurationHumanOut{
				ConfigName:  config.GetName(),
				ConfigValue: config.GetValue(),
				ReadOnly:    config.GetReadOnly(),
				Sensitive:   config.GetSensitive(),
				Source:      config.GetSource(),
				Synonyms:    strings.Join(config.GetSynonyms(), ", "),
			})
		} else {
			list.Add(&linkConfigurationSerializedOut{
				ConfigName:  config.GetName(),
				ConfigValue: config.GetValue(),
				ReadOnly:    config.GetReadOnly(),
				Sensitive:   config.GetSensitive(),
				Source:      config.GetSource(),
				Synonyms:    config.GetSynonyms(),
			})
		}
	}
	return list.Print()
}
