package plugin

import (
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

type humanRow struct {
	pluginName string
	filePath   string
}

type structuredRow struct {
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
	if err != nil {
		return err
	}
	var pluginList []row
	var overshadowedList []string
	for name, pathList := range pluginMap {
		pluginList = append(pluginList, row{pluginName: name, filePath: pathList[0]})
		for i := 1; i < len(pathList); i++ {
			overshadowedList = append(overshadowedList, pathList[i])
		}
	}

	printTable(cmd, pluginList)
	for _, path := range overshadowedList {
		utils.ErrPrintf(cmd, "	- warning: %s is overshadowed by a similarly named plugin\n", path)
	}
	return nil
}

func printTable(cmd *cobra.Command, rows []row) error {
	output, _ := cmd.Flags().GetString("output")

	w, err := poutput.NewListOutputCustomizableWriter(cmd, listFields, humanLabels, structuredLabels, cmd.OutOrStdout())
	if err != nil {
		return err
	}

	for _, row := range rows {
		if output == poutput.Human.String() {
			w.AddElement(&humanRow{
				pluginName: row.pluginName,
				filePath:   row.filePath,
			})
		} else {
			w.AddElement(&structuredRow{
				pluginName: row.pluginName,
				filePath:   row.filePath,
			})
		}
	}

	w.StableSort()
	return w.Out()
}
