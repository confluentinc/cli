package plugin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Confluent plugins.",
		Long: `Brief description:
	Confluent plugins are standalone executable files that begin with ` + "`confluent-`" + ` which are located
	on the user's $PATH.

Making plugins discoverable by the CLI:
	Plugins determine their CLI command path that they will implement based on their filenames. 
	Every sub-command in the command path that a plugin targets is separated by a dash (-). 
	For example, the plugin invoked when the user invokes the command ` + "`confluent do plugin stuff`" + ` would 
	have the filename of ` + "`confluent-do-plugin-stuff`" + `.
	Additionally, the plugin file must be moved to anywhere on the user's $PATH to be discoverable by the CLI.

Naming collisions with existing CLI commands and other plugins:
	A plugin will be overshadowed by an official existing confluent command, and thus be ignored, if and 
	only if the plugin's entire command name matches an existing official command’s name exactly. For 
	example, if a plugin is named ` + "`confluent-kafka-cluster-list`" + ` and the user invokes the plugin using 
	` + "`confluent kafka cluster list`" + `, the plugin is ignored and instead, the official command is run and a 
	warning that the plugin has been ignored is logged.  However, partial overlap between a plugin’s name 
	and an official CLI command’s name is allowed.
	If two or more plugins with the same name are found in the user's $PATH, the first one found in the $PATH
	is given precedence. The second through nth plugins discovered with the same name will be ignored.`,
	}

	c := &command{pcmd.NewAnonymousCLICommand(cmd, prerunner)}
	cmd.AddCommand(newListCommand())
	return c.Command
}
