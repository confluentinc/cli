package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *pluginCommand) newDeleteVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more custom Connect plugin versions.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.deleteVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a specific version of a custom connect plugin.",
				Code: "confluent ccpm plugin version delete ver-789012 --plugin plugin-123456 --environment env-abcdef",
			},
			examples.Example{
				Text: "Force delete a plugin version without confirmation.",
				Code: "confluent ccpm plugin version delete ver-789012 --plugin plugin-123456 --environment env-abcdef --force",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *pluginCommand) deleteVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeCCPMPluginVersion(pluginId, id, environment)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CCPMCustomConnectorPluginVersion); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCCPMPluginVersion(pluginId, id, environment)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.CCPMCustomConnectorPluginVersion)
	return err
}
