package plugin

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List plugins in user's $PATH.",
		Long:  `Lists plugins in user's $PATH; plugins are executable files that begin with "confluent-".`,
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
	// TODO: Print warnings for plugins with same name as existing commands
	for _, v := range pluginMap {
		var firstPlugin string
		for i, e := range v {
			cmd.Printf("%s\n", e)
			if i != 0 {
				cmd.Printf("	- warning: %s is overshadowed by a by a similarly named plugin: %s\n", e, firstPlugin)
			} else {
				firstPlugin = e
			}
		}
	}
	return nil
}
