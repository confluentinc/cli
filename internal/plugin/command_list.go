package plugin

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plugin"
	"github.com/confluentinc/cli/v4/pkg/types"
)

type out struct {
	Id       string `human:"ID" serialized:"id"`
	Name     string `human:"Name" serialized:"name"`
	FilePath string `human:"File Path" serialized:"file_path"`
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
	if c.cfg.DisablePlugins {
		return errors.NewErrorWithSuggestions("plugins are disabled", "To enable plugins, use `confluent configuration update disable_plugins false`.")
	}

	pluginMap := plugin.SearchPath(c.cfg)

	if len(pluginMap) == 0 && output.GetFormat(cmd) == output.Human {
		output.ErrPrintln(c.Config.EnableColor, "Please run `confluent plugin -h` for information on how to make plugins discoverable by the CLI.")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	list := output.NewList(cmd)
	var overshadowedPlugins, nameConflictPlugins []*out
	for id, paths := range pluginMap {
		path := paths[0]
		if home != "" && strings.HasPrefix(path, home) {
			path = filepath.Join("~", strings.TrimPrefix(path, home))
		}

		pluginInfo := &out{
			Name:     plugin.ToCommandName(id),
			Id:       id,
			FilePath: path,
		}

		args := strings.Split(pluginInfo.Name, " ")
		if cmd, _, _ := cmd.Root().Find(args[1:]); cmd.CommandPath() == pluginInfo.Name {
			nameConflictPlugins = append(nameConflictPlugins, pluginInfo)
		} else {
			list.Add(pluginInfo)
		}

		visitedPaths := types.NewSet(paths[0])
		for _, path := range paths[1:] {
			if visitedPaths.Contains(path) {
				continue
			}
			overshadowedPlugins = append(overshadowedPlugins, &out{Name: pluginInfo.Name, FilePath: path})
			visitedPaths.Add(path)
		}
	}

	if err := list.Print(); err != nil {
		return err
	}

	for _, pluginInfo := range nameConflictPlugins {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] The built-in command `%s` will be run instead of the duplicate plugin at %s.\n", pluginInfo.Name, pluginInfo.FilePath)
	}
	for _, pluginInfo := range overshadowedPlugins {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] The command `%s` will run the plugin listed above instead of the duplicate plugin at %s.\n", pluginInfo.Name, pluginInfo.FilePath)
	}

	return nil
}
