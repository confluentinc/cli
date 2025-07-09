package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
)

type versionOut struct {
	PluginId                  string   `human:"Plugin ID" serialized:"plugin_id"`
	PluginName                string   `human:"Plugin Name" serialized:"plugin_name"`
	Id                        string   `human:"ID" serialized:"id"`
	Version                   string   `human:"Version" serialized:"version"`
	ContentFormat             string   `human:"Content Format" serialized:"content_format"`
	DocumentationLink         string   `human:"Documentation Link" serialized:"documentation_link"`
	SensitiveConfigProperties []string `human:"Sensitive Config Properties" serialized:"sensitive_config_properties"`
	ConnectorClasses          string   `human:"Connector Classes" serialized:"connector_classes"`
	Phase                     string   `human:"Phase" serialized:"phase"`
	ErrorMessage              string   `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
	Environment               string   `human:"Environment" serialized:"environment"`
}

func (c *pluginCommand) newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage custom Connect plugin versions.",
	}

	cmd.AddCommand(c.newCreateVersionCommand())
	cmd.AddCommand(c.newDescribeVersionCommand())
	cmd.AddCommand(c.newDeleteVersionCommand())
	cmd.AddCommand(c.newListVersionCommand())

	return cmd
}

func (c *pluginCommand) printVersionTable(cmd *cobra.Command,
	plugin ccpmv1.CcpmV1CustomConnectPlugin, version ccpmv1.CcpmV1CustomConnectPluginVersion) error {
	table := output.NewTable(cmd)
	table.Add(&versionOut{
		PluginId:                  plugin.GetId(),
		PluginName:                plugin.Spec.GetDisplayName(),
		Id:                        version.GetId(),
		Version:                   version.Spec.GetVersion(),
		ContentFormat:             version.Spec.GetContentFormat(),
		DocumentationLink:         version.Spec.GetDocumentationLink(),
		SensitiveConfigProperties: version.Spec.GetSensitiveConfigProperties(),
		Phase:                     version.Status.GetPhase(),
		ConnectorClasses:          getConnectorClassesString(version.Spec.GetConnectorClasses()),
		ErrorMessage:              version.Status.GetErrorMessage(),
		Environment:               version.Spec.GetEnvironment().Id,
	})
	return table.Print()
}
