package shell

import (
	"fmt"
	goprompt "github.com/c-bata/go-prompt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/quit"
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
	cliName := c.config.CLIName

	// remove shell command from the shell
	c.RootCmd.RemoveCommand(c.Command)

	// add shell only quit command
	c.RootCmd.AddCommand(quit.NewQuitCmd(c.config))

	msg := "You are already authenticated."
	if !c.config.HasLogin() {
		msg = "You are currently not authenticated."
	}

	// run the shell
	fmt.Printf("Welcome to the %s shell! %s\n", cliName, msg)
	fmt.Println("Please press `Ctrl-D` or type `quit` to exit.")

	livePrefixFunc := func() (prefix string, useLivePrefix bool) {
		indicator := "❌ "
		if c.config.HasLogin() {
			indicator = "✅ "
		}

		return fmt.Sprintf("%s %s> ", cliName, indicator), true
	}
	livePrefixOpt := goprompt.OptionLivePrefix(livePrefixFunc)

	opts := append(prompt.DefaultPromptOptions(), livePrefixOpt)
	masterCompleter := completer.NewShellCompleter(c.RootCmd, c.config.CLIName)
	cliPrompt := prompt.NewShellPrompt(c.RootCmd, masterCompleter, c.config, opts...)
	cliPrompt.Run()
}
