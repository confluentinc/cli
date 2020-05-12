package shell

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/shell/prompt"
)

type command struct {
	Command *cobra.Command
	RootCmd *cobra.Command
	config  *v3.Config
}

// NewShellCmd returns the Cobra command for the shell.
func NewShellCmd(rootCmd *cobra.Command, config *v3.Config) *cobra.Command {
	cliCmd := &command{
		RootCmd: rootCmd,
		config:  config,
	}

	cliCmd.init()
	return cliCmd.Command
}

func (c *command) init() {
	c.Command = &cobra.Command{
		Use:   "shell",
		Short: fmt.Sprintf("Run the %s shell.", c.config.CLIName),
		Run:   c.shell,
		Args:  cobra.NoArgs,
	}
}

func (c *command) shell(cmd *cobra.Command, args []string) {
	// remove shell command from the shell
	c.RootCmd.RemoveCommand(c.Command)

	// run the shell
	fmt.Printf("Welcome to the %s shell!\n", c.config.CLIName)
	fmt.Println("Please press `Ctrl-D` to exit.")
	masterCompleter := completer.NewShellCompleter(c.RootCmd, c.config.CLIName)
	cliPrompt := prompt.NewShellPrompt(c.RootCmd, masterCompleter, c.config, prompt.DefaultPromptOptions()...)
	cliPrompt.Run()
}
