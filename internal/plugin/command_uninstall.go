package plugin

import (
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plugin"
	"github.com/confluentinc/cli/v4/pkg/resource"
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

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.Plugin); err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		return os.Remove(pluginMap[name][0])
	}

	if _, err := deletion.Delete(cmd, args, deleteFunc, resource.Plugin); err != nil {
		return err
	}

	if c.cfg.DisablePlugins {
		output.ErrPrintln(c.Config.EnableColor, disabledPluginsWarning)
	}

	return nil
}
