package plugin

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"sort"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List plugins in user's $PATH.",
		Long:  `List Confluent CLI plugins in user's $PATH. Plugins are executable files that begin with "confluent-".`,
		Args:  cobra.NoArgs,
		RunE:  list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func list(cmd *cobra.Command, _ []string) error {
	pluginMap, err := plugin.SearchPath()
	var pluginList []string
	if err != nil {
		return err
	}
	for _, v := range pluginMap {
		var firstPlugin string
		for i, e := range v {
			pluginList = append(pluginList, e)
			if i != 0 {
				utils.ErrPrintf(cmd, "	- warning: %s is overshadowed by a similarly named plugin: %s\n", e, firstPlugin)
			} else {
				firstPlugin = e
			}
		}
	}
	sort.Strings(pluginList)
	for _, v := range pluginList {
		fmt.Println(v)
	}
	return nil
}
