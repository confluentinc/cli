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
		Use:   "delete <version-id-1> [version-id-2] ... [version-id-n]",
		Short: "Delete one or more Custom Connect Plugin Versions.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.deleteVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a specific version of a custom connect plugin.",
				Code: "confluent ccpm plugin version delete version-789012 --plugin plugin-123456 --environment env-abcdef",
			},
			examples.Example{
				Text: "Delete multiple versions of a plugin.",
				Code: "confluent ccpm plugin version delete version-1.0.0 version-2.0.0 --plugin plugin-123456 --environment env-abcdef",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	cmd.Flags().String("environment", "", "Environment ID.")
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *pluginCommand) deleteVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
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

	_, err = deletion.Delete(args, deleteFunc, resource.CCPMCustomConnectorPluginVersion)
	return err
}
