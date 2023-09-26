package connect

import (
	"strings"

	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type customPluginCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type customPluginSerializedOut struct {
	Id                  string   `serialized:"id"`
	Name                string   `serialized:"name"`
	Description         string   `serialized:"description"`
	ConnectorClass      string   `serialized:"connector_class"`
	ConnectorType       string   `serialized:"connector_type"`
	SensitiveProperties []string `serialized:"sensitive_properties"`
}

type customPluginHumanOut struct {
	Id                  string `human:"ID"`
	Name                string `human:"Name"`
	Description         string `human:"Description"`
	ConnectorClass      string `human:"Connector Class"`
	ConnectorType       string `human:"Connector Type"`
	SensitiveProperties string `human:"Sensitive Properties"`
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
	sensitiveProperties := plugin.GetSensitiveConfigProperties()
	if output.GetFormat(cmd) == output.Human {
		table.Add(&customPluginHumanOut{
			Id:                  plugin.GetId(),
			Name:                plugin.GetDisplayName(),
			Description:         plugin.GetDescription(),
			ConnectorClass:      plugin.GetConnectorClass(),
			ConnectorType:       plugin.GetConnectorType(),
			SensitiveProperties: strings.Join(sensitiveProperties, ","),
		})
	} else {
		table.Add(&customPluginSerializedOut{
			Id:                  plugin.GetId(),
			Name:                plugin.GetDisplayName(),
			Description:         plugin.GetDescription(),
			ConnectorClass:      plugin.GetConnectorClass(),
			ConnectorType:       plugin.GetConnectorType(),
			SensitiveProperties: sensitiveProperties,
		})
	}

	return table.Print()
}
