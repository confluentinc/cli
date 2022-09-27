package plugin

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type out struct {
	PluginName string `human:"Plugin Name" serialized:"plugin_name"`
	FilePath   string `human:"File Path" serialized:"file_path"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent CLI plugins in $PATH.",
		Long:  `List Confluent CLI plugins in $PATH. Plugins are executable files that begin with "confluent-".`,
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	pluginMap, err := plugin.SearchPath(c.cfg)
	if err != nil {
		return err
	}

	if len(pluginMap) == 0 && output.GetFormat(cmd) == output.Human {
		utils.ErrPrintln(cmd, "Please run `confluent plugin -h` for information on how to make plugins discoverable by the CLI.")
	}

	list := output.NewList(cmd)
	var overshadowedPlugins, nameConflictPlugins []*out
	for name, pathList := range pluginMap {
		pluginInfo := &out{
			PluginName: strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", "-"),
			FilePath:   pathList[0],
		}
		args := strings.Split(pluginInfo.PluginName, " ")
		if cmd, _, _ := cmd.Root().Find(args[1:]); cmd.CommandPath() == pluginInfo.PluginName {
			nameConflictPlugins = append(nameConflictPlugins, pluginInfo)
		} else {
			list.Add(pluginInfo)
		}
		for i := 1; i < len(pathList); i++ {
			overshadowedPlugins = append(overshadowedPlugins, &out{PluginName: pluginInfo.PluginName, FilePath: pathList[i]})
		}
	}

	if err := list.Print(); err != nil {
		return err
	}

	for _, pluginInfo := range nameConflictPlugins {
		utils.ErrPrintf(cmd, "[WARN] The built-in command `%s` will be run instead of the duplicate plugin at %s.\n", pluginInfo.PluginName, pluginInfo.FilePath)
	}
	for _, pluginInfo := range overshadowedPlugins {
		utils.ErrPrintf(cmd, "[WARN] The command `%s` will run the plugin listed above instead of the duplicate plugin at %s.\n", pluginInfo.PluginName, pluginInfo.FilePath)
	}

	return nil
}
