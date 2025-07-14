package connect

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type customPluginCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type customPluginOut struct {
	Id                  string   `human:"ID" serialized:"id"`
	Name                string   `human:"Name" serialized:"name"`
	Description         string   `human:"Description" serialized:"description"`
	Cloud               string   `human:"Cloud" serialized:"cloud"`
	ConnectorClass      string   `human:"Connector Class" serialized:"connector_class"`
	ConnectorType       string   `human:"Connector Type" serialized:"connector_type"`
	SensitiveProperties []string `human:"Sensitive Properties" serialized:"sensitive_properties"`
}

func newCustomPluginCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "custom-plugin",
		Short:       "Manage custom connector plugins.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &customPluginCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printTable(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin) error {
	table := output.NewTable(cmd)
	table.Add(&customPluginOut{
		Id:                  plugin.GetId(),
		Name:                plugin.GetDisplayName(),
		Description:         plugin.GetDescription(),
		Cloud:               plugin.GetCloud(),
		ConnectorClass:      plugin.GetConnectorClass(),
		ConnectorType:       plugin.GetConnectorType(),
		SensitiveProperties: plugin.GetSensitiveConfigProperties(),
	})
	return table.Print()
}
