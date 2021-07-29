package shell

import (
	"fmt"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/quit"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/shell/prompt"
	"github.com/confluentinc/cli/internal/pkg/version"
)

const (
	watermelonRed   = goprompt.Color(167)
	candyAppleGreen = goprompt.Color(77)
)

type command struct {
	Command      *cobra.Command
	RootCmd      *cobra.Command
	config       *v3.Config
	prerunner    pcmd.PreRunner
	completer    *completer.ShellCompleter
	jwtValidator pcmd.JWTValidator
}

// NewShellCmd returns the Cobra command for the shell.
func NewShellCmd(rootCmd *cobra.Command, prerunner pcmd.PreRunner, config *v3.Config, completer *completer.ShellCompleter, jwtValidator pcmd.JWTValidator) *cobra.Command {
	cliCmd := &command{
		RootCmd:      rootCmd,
		config:       config,
		prerunner:    prerunner,
		completer:    completer,
		jwtValidator: jwtValidator,
	}

	cliCmd.init()
	return cliCmd.Command
}

func (c *command) init() {
	c.Command = &cobra.Command{
		Use:   "shell",
		Short: fmt.Sprintf("Run the %s shell.", version.CLIName),
		RunE:  pcmd.NewCLIRunE(c.shell),
		Args:  cobra.NoArgs,
	}
}

func (c *command) shell(cmd *cobra.Command, args []string) error {
	// remove shell command from the shell
	c.RootCmd.RemoveCommand(c.Command)

	// add shell only quit command
	c.RootCmd.AddCommand(quit.NewQuitCmd(c.prerunner))

	msg := errors.AlreadyAuthenticatedMsg
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	// For the first time, validate the token using the prerunner, which tries to update the JWT if it's invalid.
	// After the first time, validate using the token validator, which doesn't try to update the JWT because that would be too slow.
	// Also, let the prerunner track the command analytics as usual, since the shell command doesn't have a normal prerunner that would do this.
	// TODO: Make the command an AnonymousCLICommand and clean this up to just use the JWT validator.
	if err := c.prerunner.Authenticated(pcmd.NewAuthenticatedCLICommand(c.Command, c.prerunner))(c.Command, args); err != nil {
		msg = errors.CurrentlyNotAuthenticatedMsg
	}

	// run the shell
	fmt.Printf(errors.ShellWelcomeMsg, version.CLIName, msg)
	fmt.Println(errors.ShellExitInstructionsMsg)

	opts := prompt.DefaultPromptOptions()
	cliPrompt := prompt.NewShellPrompt(c.RootCmd, c.completer, opts...)
	livePrefixOpt := goprompt.OptionLivePrefix(livePrefixFunc(cliPrompt.Prompt, c.config, c.jwtValidator))
	if err := livePrefixOpt(cliPrompt.Prompt); err != nil {
		// This returns nil in the go-prompt implementation.
		// This is also what go-prompt does if err != nil.
		panic(err)
	}
	cliPrompt.Run()
	return nil
}

func livePrefixFunc(prompt *goprompt.Prompt, config *v3.Config, jwtValidator pcmd.JWTValidator) func() (prefix string, useLivePrefix bool) {
	return func() (prefix string, useLivePrefix bool) {
		text, color := prefixState(jwtValidator, config)
		if err := goprompt.OptionPrefixTextColor(color)(prompt); err != nil {
			// This returns nil in the go-prompt implementation.
			// This is also what go-prompt does if err != nil for all of its options.
			panic(err)
		}
		return text, true
	}
}

// prefixState returns the text and color of the prompt prefix.
func prefixState(jwtValidator pcmd.JWTValidator, config *v3.Config) (text string, color goprompt.Color) {
	prefixColor := watermelonRed
	if err := jwtValidator.Validate(config.Context()); err == nil {
		prefixColor = candyAppleGreen
	}
	return fmt.Sprintf("%s > ", version.CLIName), prefixColor
}
