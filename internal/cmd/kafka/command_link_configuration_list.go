package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type linkConfigurationOut struct {
	ConfigName  string   `human:"Config Name" serialized:"config_name"`
	ConfigValue string   `human:"Config Value" serialized:"config_value"`
	ReadOnly    bool     `human:"Read-Only" serialized:"read_only"`
	Sensitive   bool     `human:"Sensitive" serialized:"sensitive"`
	Source      string   `human:"Source" serialized:"source"`
	Synonyms    []string `human:"Synonyms" serialized:"synonyms"`
}

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
	linkName := args[0]

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	configs, err := kafkaREST.CloudClient.ListKafkaLinkConfigs(linkName)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	list.Add(&linkConfigurationOut{
		ConfigName:  "dest.cluster.id",
		ConfigValue: kafkaREST.GetClusterId(),
		ReadOnly:    true,
		Sensitive:   true,
	})

	for _, config := range configs.GetData() {
		list.Add(&linkConfigurationOut{
			ConfigName:  config.GetName(),
			ConfigValue: config.GetValue(),
			ReadOnly:    config.GetReadOnly(),
			Sensitive:   config.GetSensitive(),
			Source:      config.GetSource(),
			Synonyms:    config.GetSynonyms(),
		})
	}
	return list.Print()
}
