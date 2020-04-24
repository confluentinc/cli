package shell

import (
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/shell/prompt"
)

// NewShellCmd returns the Cobra command for the shell.
func NewShellCmd(prerunner pcmd.PreRunner, rootCmd *cobra.Command, config *v3.Config) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "shell",
			Short: "Run the " + config.CLIName + " shell.",
			Run: func(cmd *cobra.Command, args []string) {
				// remove shell command from the shell
				rootCmd.RemoveCommand(cmd)

				// run the shell
				fmt.Println("Welcome to the " + config.CLIName + " shell!")
				fmt.Println("Please press ctrl-D to exit.")
				masterCompleter := completer.NewShellCompleter(rootCmd, config.CLIName)
				cliPrompt := prompt.NewShellPrompt(rootCmd, masterCompleter, config, prompt.DefaultPromptOptions()...)
				cliPrompt.Run()
			},
			Args: cobra.NoArgs,
		},
		config, prerunner)
	return cliCmd.Command
}
