package plugin

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/plugin"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/spf13/cobra"
	"os"
)

func (c *command) newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <plugin-id-1> [plugin-id-2] ... [plugin-id-n]",
		Short: "Uninstall official Confluent CLI plugins.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.uninstall,
	}
}

func (c *command) uninstall(cmd *cobra.Command, args []string) error {
	pcmd.AddForceFlag(cmd)
	pluginMap := plugin.SearchPath(c.cfg)
	existenceFunc := func(name string) bool {
		_, ok := pluginMap[name]
		return ok
	}

	if err := deletion.ValidateAndConfirmDeletionYesNo(cmd, args, existenceFunc, resource.Plugin); err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		if err := os.Remove(pluginMap[name][0]); err != nil {
			return err
		}
		return nil
	}

	_, err := deletion.Delete(args, deleteFunc, resource.Plugin)
	return err
}
