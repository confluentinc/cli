package shell

import (
	"fmt"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/quit"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/shell/prompt"
)

const (
	watermelonRed   = goprompt.Color(167)
	candyAppleGreen = goprompt.Color(77)
)

type command struct {
	Command   *cobra.Command
	RootCmd   *cobra.Command
	config    *v3.Config
	prerunner pcmd.PreRunner
	completer *completer.ShellCompleter
}

// NewShellCmd returns the Cobra command for the shell.
func NewShellCmd(rootCmd *cobra.Command, config *v3.Config, prerunner pcmd.PreRunner, completer *completer.ShellCompleter) *cobra.Command {
	cliCmd := &command{
		RootCmd:   rootCmd,
		config:    config,
		prerunner: prerunner,
		completer: completer,
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
	if err := c.prerunner.Authenticated(pcmd.NewAuthenticatedCLICommand(c.Command, c.prerunner))(c.Command, args); err != nil {
		msg = "You are currently not authenticated."
	}

	// run the shell
	fmt.Printf("Welcome to the %s shell! %s\n", cliName, msg)
	fmt.Println("Please press `Ctrl-D` or type `quit` to exit.")

	opts := prompt.DefaultPromptOptions()
	cliPrompt := prompt.NewShellPrompt(c.RootCmd, c.completer, c.config, opts...)
	livePrefixOpt := goprompt.OptionLivePrefix(livePrefixFunc(cliPrompt))
	if err := livePrefixOpt(cliPrompt.Prompt); err != nil {
		// This returns nil in the go-prompt implementation.
		// This is also what go-prompt does if err != nil.
		panic(err)
	}
	cliPrompt.Run()
}

func livePrefixFunc(cliPrompt *prompt.ShellPrompt) func() (prefix string, useLivePrefix bool) {
	return func() (prefix string, useLivePrefix bool) {
		text, color := prefixState(cliPrompt.Config)
		if err := goprompt.OptionPrefixTextColor(color)(cliPrompt.Prompt); err != nil {
			// This returns nil in the go-prompt implementation.
			// This is also what go-prompt does if err != nil for all of its options.
			panic(err)
		}
		return text, true
	}
}

// prefixState returns the text and color of the prompt prefix.
func prefixState(config *v3.Config) (text string, color goprompt.Color) {
	prefixColor := watermelonRed
	if config.HasLogin() {
		prefixColor = candyAppleGreen
	}
	return fmt.Sprintf("%s > ", config.CLIName), prefixColor
}
