package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type linkConfigurationOut struct {
	ConfigName  string   `human:"Config Name" serialized:"config_name"`
	ConfigValue string   `human:"Config Value" serialized:"config_value"`
	ReadOnly    bool     `human:"Read Only" serialized:"read_only"`
	Sensitive   bool     `human:"Sensitive" serialized:"sensitive"`
	Source      string   `human:"Source" serialized:"source"`
	Synonyms    []string `human:"Synonyms" serialized:"synonyms"`
}

func (c *linkCommand) newConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <name>",
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
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	clusterId, err := getKafkaClusterLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	listLinkConfigsRespData, httpResp, err := kafkaREST.CloudClient.ListKafkaLinkConfigs(clusterId, linkName)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
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
