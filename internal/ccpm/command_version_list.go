package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *pluginCommand) newListVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom Connect plugin versions.",
		Args:  cobra.NoArgs,
		RunE:  c.listVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all versions of a custom connect plugin.",
				Code: "confluent ccpm plugin version list --plugin plugin-123456 --environment env-abcdef",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *pluginCommand) listVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pluginResp, err := c.V2Client.DescribeCCPMPlugin(pluginId, environment)
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	versions, err := c.V2Client.ListCCPMPluginVersions(pluginId, environment)
	if err != nil {
		return err
	}

	// Display results in table format
	table := output.NewList(cmd)
	for _, version := range versions {
		spec, _ := version.GetSpecOk()
		status, _ := version.GetStatusOk()

		table.Add(&versionOut{
			PluginId:                  pluginResp.GetId(),
			PluginName:                pluginResp.Spec.GetDisplayName(),
			Id:                        version.GetId(),
			Version:                   spec.GetVersion(),
			ContentFormat:             spec.GetContentFormat(),
			DocumentationLink:         spec.GetDocumentationLink(),
			SensitiveConfigProperties: spec.GetSensitiveConfigProperties(),
			ConnectorClasses:          getConnectorClassesString(spec.GetConnectorClasses()),
			Phase:                     status.GetPhase(),
			ErrorMessage:              status.GetErrorMessage(),
			Environment:               spec.Environment.Id,
		})
	}

	table.Sort(true)
	return table.Print()
}
