package plugin

import (
	"sort"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

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
	var pluginList []string
	for _, pluginNames := range pluginMap {
		pluginList = append(pluginList, pluginNames[0])
		for i := 1; i < len(pluginNames); i++ {
			pluginList = append(pluginList, pluginNames[i])
			utils.ErrPrintf(cmd, "	- warning: %s is overshadowed by a similarly named plugin: %s\n", pluginNames[i], pluginNames[0])
		}
	}
	sort.Strings(pluginList)
	for _, pluginPath := range pluginList {
		utils.Println(cmd, pluginPath)
	}
	return nil
}
