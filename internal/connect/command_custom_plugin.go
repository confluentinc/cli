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
	Cloud               string   `serialized:"cloud"`
	ConnectorClass      string   `serialized:"connector_class"`
	ConnectorType       string   `serialized:"connector_type"`
	SensitiveProperties []string `serialized:"sensitive_properties"`
}

type customPluginHumanOut struct {
	Id                  string `human:"ID"`
	Name                string `human:"Name"`
	Description         string `human:"Description"`
	Cloud               string `human:"Cloud"`
	ConnectorClass      string `human:"Connector Class"`
	ConnectorType       string `human:"Connector Type"`
	SensitiveProperties string `human:"Sensitive Properties"`
}

type customPluginVersionSerializedOut struct {
	Plugin              string   `serialized:"plugin"`
	Version             string   `serialized:"version"`
	Name                string   `serialized:"name"`
	Description         string   `serialized:"description"`
	Cloud               string   `serialized:"cloud"`
	ConnectorClass      string   `serialized:"connector_class"`
	ConnectorType       string   `serialized:"connector_type"`
	SensitiveProperties []string `serialized:"sensitive_properties"`
	VersionNumber       string   `serialized:"version_number"`
	IsBeta              string   `serialized:"is_beta"`
	ReleaseNotes        string   `serialized:"release_notes"`
}

type customPluginVersionHumanOut struct {
	PluginId            string `human:"Plugin"`
	Version             string `human:"Version"`
	Name                string `human:"Name"`
	Description         string `human:"Description"`
	Cloud               string `human:"Cloud"`
	ConnectorClass      string `human:"Connector Class"`
	ConnectorType       string `human:"Connector Type"`
	SensitiveProperties string `human:"Sensitive Properties"`
	VersionNumber       string `human:"Version Number"`
	IsBeta              string `human:"Is Beta"`
	ReleaseNotes        string `human:"Release Notes"`
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
	cmd.AddCommand(c.newCreateVersionCommand())
	cmd.AddCommand(c.newDescribeVersionCommand())
	cmd.AddCommand(c.newListVersionCommand())
	cmd.AddCommand(c.newDeleteVersionCommand())
	cmd.AddCommand(c.newUpdateVersionCommand())

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
			Cloud:               plugin.GetCloud(),
			ConnectorClass:      plugin.GetConnectorClass(),
			ConnectorType:       plugin.GetConnectorType(),
			SensitiveProperties: strings.Join(sensitiveProperties, ", "),
		})
	} else {
		table.Add(&customPluginSerializedOut{
			Id:                  plugin.GetId(),
			Name:                plugin.GetDisplayName(),
			Description:         plugin.GetDescription(),
			Cloud:               plugin.GetCloud(),
			ConnectorClass:      plugin.GetConnectorClass(),
			ConnectorType:       plugin.GetConnectorType(),
			SensitiveProperties: sensitiveProperties,
		})
	}

	return table.Print()
}

func printTableVersion(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin, version connectcustompluginv1.ConnectV1CustomConnectorPluginVersion) error {
	table := output.NewTable(cmd)
	sensitiveProperties := plugin.GetSensitiveConfigProperties()
	if output.GetFormat(cmd) == output.Human {
		table.Add(&customPluginVersionHumanOut{
			PluginId:            plugin.GetId(),
			Name:                plugin.GetDisplayName(),
			Description:         plugin.GetDescription(),
			Cloud:               plugin.GetCloud(),
			ConnectorClass:      plugin.GetConnectorClass(),
			ConnectorType:       plugin.GetConnectorType(),
			SensitiveProperties: strings.Join(sensitiveProperties, ", "),
			Version:             version.GetId(),
			VersionNumber:       version.GetVersion(),
			IsBeta:              version.GetIsBeta(),
			ReleaseNotes:        version.GetReleaseNotes(),
		})
	} else {
		table.Add(&customPluginVersionSerializedOut{
			Plugin:              plugin.GetId(),
			Name:                plugin.GetDisplayName(),
			Description:         plugin.GetDescription(),
			Cloud:               plugin.GetCloud(),
			ConnectorClass:      plugin.GetConnectorClass(),
			ConnectorType:       plugin.GetConnectorType(),
			SensitiveProperties: sensitiveProperties,
			Version:             version.GetId(),
			VersionNumber:       version.GetVersion(),
			IsBeta:              version.GetIsBeta(),
			ReleaseNotes:        version.GetReleaseNotes(),
		})
	}

	return table.Print()
}
