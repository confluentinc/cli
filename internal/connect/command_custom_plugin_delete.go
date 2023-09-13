package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *customPluginCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete custom connector plugin.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a custom connector plugin",
				Code: "confluent connect custom-plugin delete ccp-123456",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *customPluginCommand) delete(cmd *cobra.Command, args []string) error {
	pluginIdToName, err := c.mapPluginIdToName()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, ok := pluginIdToName[id]
		return ok
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.CustomConnectorPlugin, pluginIdToName[args[0]]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCustomPlugin(id)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.CustomConnectorPlugin)
	return err
}

func (c *customPluginCommand) mapPluginIdToName() (map[string]string, error) {
	plugins, err := c.V2Client.ListCustomPlugins()
	if err != nil {
		return nil, err
	}

	pluginIdToName := make(map[string]string)
	for _, plugin := range plugins.GetData() {
		pluginIdToName[plugin.GetId()] = plugin.GetDisplayName()
	}

	return pluginIdToName, nil
}
