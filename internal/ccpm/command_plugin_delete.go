package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *pluginCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more custom Connect plugins.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a custom Connect plugin by ID.",
				Code: "confluent ccpm plugin delete plugin-123456 --environment env-12345",
			},
			examples.Example{
				Text: "Force delete a custom Connect plugin without confirmation.",
				Code: "confluent ccpm plugin delete plugin-123456 --environment env-12345 --force",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *pluginCommand) delete(cmd *cobra.Command, args []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	cloud := ""

	existenceFunc := func(id string) bool {
		customConnectPluginIdToName, err := c.mapCustomConnectPluginIdToName(cloud, environment)
		if err != nil {
			return false
		}
		_, ok := customConnectPluginIdToName[id]
		return ok
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CCPMCustomConnectorPlugin); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		// Use V2Client to call CCPM API
		return c.V2Client.DeleteCCPMPlugin(id, environment)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.CCPMCustomConnectorPlugin)
	return err
}

func (c *pluginCommand) mapCustomConnectPluginIdToName(cloud, environment string) (map[string]string, error) {
	plugins, err := c.V2Client.ListCCPMPlugins(cloud, environment)
	if err != nil {
		return nil, err
	}

	customConnectPluginIdToName := make(map[string]string, len(plugins))
	for _, plugin := range plugins {
		customConnectPluginIdToName[*plugin.Id] = *plugin.Spec.DisplayName
	}
	return customConnectPluginIdToName, nil
}
