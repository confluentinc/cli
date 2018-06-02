package common

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const longDescription = `Output shell completion code for the specified shell (bash or zsh).
The shell code must be evaluated to provide interactive completion of confluent commands.

Install bash completions on macOS:

  # Enable bash completions using homebrew
  ## If running Bash 3.2 included with macOS
  brew install bash-completion
  ## or, if running Bash 4.1+
  brew install bash-completion@2
  # Set the confluent completion code for bash to a file that's sourced on login
  confluent completion bash > $(brew --prefix)/etc/bash_completion.d/confluent

Install bash completions on Linux:

  # Load the confluent completion code for bash into the current shell
  source <(confluent completion bash)

  # Set the confluent completion code for bash to a file that's sourced on login
  confluent completion bash > /etc/bash_completion.d/confluent

Install zsh completions:

  # Load the confluent completion code for zsh into the current shell
  source <(confluent completion zsh)

  # Set the confluent completion code for zsh to autoload on startup
  confluent completion zsh > "${fpath[1]}/_confluent"
`

type command struct {
	*cobra.Command
	rootCmd *cobra.Command
}

// NewCompletionCmd returns the Cobra command for shell completion.
func NewCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &command{
		rootCmd: rootCmd,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Command = &cobra.Command{
		Use:   "completion SHELL",
		Short: "Output shell completion code for the specified shell (bash or zsh).",
		Long:  longDescription,
		RunE:  c.completion,
		Args:  cobra.ExactArgs(1),
	}
}

func (c *command) completion(cmd *cobra.Command, args []string) error {
	var err error
	if args[0] == "bash" {
		err = c.rootCmd.GenBashCompletion(os.Stdout)
	} else if args[0] == "zsh" {
		err = c.rootCmd.GenZshCompletion(os.Stdout)
	} else {
		err = fmt.Errorf(`unsupported shell type "%s"`, args[0])
	}
	return err
}
