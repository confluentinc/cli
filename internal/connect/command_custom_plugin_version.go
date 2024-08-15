package connect

import (
	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"strings"
)

type customPluginVersionCommand struct {
	*pcmd.AuthenticatedCLICommand
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
	Plugin              string `human:"Plugin"`
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

func newCustomPluginVersionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "version",
		Short:       "Manage custom connector plugins versions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &customPluginCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateVersionCommand())
	cmd.AddCommand(c.newDescribeVersionCommand())
	cmd.AddCommand(c.newListVersionCommand())
	cmd.AddCommand(c.newDeleteVersionCommand())
	cmd.AddCommand(c.newUpdateVersionCommand())

	return cmd
}

func printTableVersion(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin, version connectcustompluginv1.ConnectV1CustomConnectorPluginVersion) error {
	table := output.NewTable(cmd)
	sensitiveProperties := plugin.GetSensitiveConfigProperties()
	if output.GetFormat(cmd) == output.Human {
		table.Add(&customPluginVersionHumanOut{
			Plugin:              plugin.GetId(),
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
