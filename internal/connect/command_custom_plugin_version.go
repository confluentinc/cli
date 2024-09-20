package connect

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type customPluginVersionOut struct {
	Plugin              string   `human:"Plugin" serialized:"plugin"`
	Version             string   `human:"Version" serialized:"version"`
	Name                string   `human:"Name" serialized:"name"`
	Description         string   `human:"Description" serialized:"description"`
	Cloud               string   `human:"Cloud" serialized:"cloud"`
	ConnectorClass      string   `human:"Connector Class" serialized:"connector_class"`
	ConnectorType       string   `human:"Connector Type" serialized:"connector_type"`
	SensitiveProperties []string `human:"Sensitive Properties" serialized:"sensitive_properties"`
	VersionNumber       string   `human:"Version Number" serialized:"version_number"`
	IsBeta              bool     `human:"Beta" serialized:"is_beta"`
	ReleaseNotes        string   `human:"Release Notes" serialized:"release_notes"`
	ErrorTrace          string   `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *customPluginCommand) newCustomPluginVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage custom connector plugin versions.",
	}

	cmd.AddCommand(c.newVersionCreateCommand())
	cmd.AddCommand(c.newVersionDeleteCommand())
	cmd.AddCommand(c.newVersionDescribeCommand())
	cmd.AddCommand(c.newVersionListCommand())
	cmd.AddCommand(c.newVersionUpdateCommand())

	return cmd
}

func printTableVersion(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin, version connectcustompluginv1.ConnectV1CustomConnectorPluginVersion) error {
	table := output.NewTable(cmd)
	table.Add(&customPluginVersionOut{
		Plugin:              plugin.GetId(),
		Name:                plugin.GetDisplayName(),
		Description:         plugin.GetDescription(),
		Cloud:               plugin.GetCloud(),
		ConnectorClass:      plugin.GetConnectorClass(),
		ConnectorType:       plugin.GetConnectorType(),
		SensitiveProperties: version.GetSensitiveConfigProperties(),
		Version:             version.GetId(),
		VersionNumber:       version.GetVersion(),
		IsBeta:              version.GetIsBeta() == "true",
		ReleaseNotes:        version.GetReleaseNotes(),
	})

	return table.Print()
}
