package plugin

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/confluentinc/cli/internal/pkg/set"
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
	pluginMap := plugin.SearchPath(c.cfg)

	if len(pluginMap) == 0 && output.GetFormat(cmd) == output.Human {
		utils.ErrPrintln(cmd, "Please run `confluent plugin -h` for information on how to make plugins discoverable by the CLI.")
	}

	list := output.NewList(cmd)
	var overshadowedPlugins, nameConflictPlugins []*out
	for name, paths := range pluginMap {
		pluginInfo := &out{
			PluginName: strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", "-"),
			FilePath:   paths[0],
		}
		args := strings.Split(pluginInfo.PluginName, " ")
		if cmd, _, _ := cmd.Root().Find(args[1:]); cmd.CommandPath() == pluginInfo.PluginName {
			nameConflictPlugins = append(nameConflictPlugins, pluginInfo)
		} else {
			list.Add(pluginInfo)
		}

		visitedPaths := set.New(paths[0])
		for _, path := range paths[1:] {
			if visitedPaths.Contains(path) {
				continue
			}
			overshadowedPlugins = append(overshadowedPlugins, &out{PluginName: pluginInfo.PluginName, FilePath: path})
			visitedPaths.Add(path)
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
