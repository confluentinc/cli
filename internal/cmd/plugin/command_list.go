package plugin

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	poutput "github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listFields       = []string{"pluginName", "filePath"}
	humanLabels      = []string{"Plugin Name", "File Path"}
	structuredLabels = []string{"plugin_name", "file_path"}
)

type row struct {
	pluginName string
	filePath   string
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent CLI plugins in $PATH.",
		Long:  `List Confluent CLI plugins in $PATH. Plugins are executable files that begin with "confluent-".`,
		Args:  cobra.NoArgs,
		RunE:  list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func list(cmd *cobra.Command, _ []string) error {
	pluginMap, err := plugin.SearchPath()
	if len(pluginMap) == 0 && cmd.Flag("output").Value.String() == "human" {
		utils.ErrPrintln(cmd, "Please run `confluent plugin -h` for information on how to make plugins discoverable by the CLI.")
	}
	if err != nil {
		return err
	}
	var pluginList, overshadowedPlugins, nameConflictPlugins []row
	for name, pathList := range pluginMap {
		pluginInfo := row{
			pluginName: strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", "-"),
			filePath:   pathList[0],
		}
		args := strings.Split(pluginInfo.pluginName, " ")
		if cmd, _, _ := cmd.Root().Find(args[1:]); cmd.CommandPath() == pluginInfo.pluginName {
			nameConflictPlugins = append(nameConflictPlugins, pluginInfo)
		} else {
			pluginList = append(pluginList, pluginInfo)
		}
		for i := 1; i < len(pathList); i++ {
			overshadowedPlugins = append(overshadowedPlugins, row{pluginName: pluginInfo.pluginName, filePath: pathList[i]})
		}
	}

	if err := printTable(cmd, pluginList); err != nil {
		return err
	}
	for _, pluginInfo := range nameConflictPlugins {
		utils.ErrPrintf(cmd, "[WARN] The built-in command `%s` will be run instead of the duplicate plugin at %s.\n", pluginInfo.pluginName, pluginInfo.filePath)
	}
	for _, pluginInfo := range overshadowedPlugins {
		utils.ErrPrintf(cmd, "[WARN] The command `%s` will run the plugin listed above instead of the duplicate plugin at %s.\n", pluginInfo.pluginName, pluginInfo.filePath)
	}
	return nil
}

func printTable(cmd *cobra.Command, rows []row) error {
	w, err := poutput.NewListOutputCustomizableWriter(cmd, listFields, humanLabels, structuredLabels, cmd.OutOrStdout())
	if err != nil {
		return err
	}

	for _, r := range rows {
		w.AddElement(&row{
			pluginName: r.pluginName,
			filePath:   r.filePath,
		})
	}

	w.StableSort()
	return w.Out()
}
