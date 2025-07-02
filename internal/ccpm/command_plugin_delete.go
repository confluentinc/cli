package ccpm

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/spf13/cobra"
)

func (c *pluginCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Custom Connect Plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	cmd.Flags().String("environment", "", "Environment ID.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *pluginCommand) delete(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
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

	_, err = deletion.Delete(args, deleteFunc, resource.CCPMCustomConnectorPlugin)
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
