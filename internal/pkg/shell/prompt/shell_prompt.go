package prompt

import (
	"strings"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

const (
	defaultShellPrefix = " > "
)

type ShellPrompt struct {
	*goprompt.Prompt
	RootCmd *cobra.Command
	completer.Completer
	Config *v3.Config
}

func NewShellPrompt(rootCmd *cobra.Command, compl completer.Completer, cfg *v3.Config, opts ...goprompt.Option) *ShellPrompt {
	shell := &ShellPrompt{
		Completer: compl,
		RootCmd:   rootCmd,
		Config:    cfg,
	}

	prompt := goprompt.New(
		func(in string) {
			promptArgs := strings.Fields(in)
			rootCmd.SetArgs(promptArgs)
			rootCmd.Execute()
		},
		shell.Complete,
		opts...,
	)
	shell.Prompt = prompt

	return shell
}

func DefaultPromptOptions() []goprompt.Option {
	return []goprompt.Option{
		goprompt.OptionShowCompletionAtStart(),
		goprompt.OptionPrefix(defaultShellPrefix),
	}
}
