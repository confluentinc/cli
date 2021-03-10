package prompt

import (
	"os"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	shellparser "mvdan.cc/sh/v3/shell"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/feedback"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

const (
	defaultShellPrefix = " > "
	maxSuggestion      = 15
)

type ShellPrompt struct {
	*goprompt.Prompt
	RootCmd *instrumentedCommand
	completer.Completer
	Config *v3.Config
}

type instrumentedCommand struct {
	*cobra.Command
	analytics analytics.Client
	logger    *log.Logger
}

// Slice and array flags need to get reset;
// Cobra/flags really weren't designed for this use case
// where flags are getting set over and over on multiple
// shell command runs, and array/slice flags seem especially
// brittle.  Without resetting these flag values, the arrays
// just get appended to ad infinitum.  Even with this fix,
// it seems like the appending is still happening somewhere
// in the internals of cobra/flags; however, this bandaid
// still seems to be enough to counteract that behavior.
//
// Note: Persistent flags need to be visited, too, if the need arises.
func resetArrayAndSliceFlags(c *cobra.Command) {
	c.Flags().VisitAll(func(f *flag.Flag) {
		if sliceValue, ok := f.Value.(flag.SliceValue); ok {
			_ := sliceValue.Replace([]string{})
		}
	})
	if c.HasSubCommands() {
		for _, y := range c.Commands() {
			resetArrayAndSliceFlags(y)
		}
	}
}

func (c *instrumentedCommand) Execute(cliName string, args []string) error {
	c.analytics.SetStartTime()
	if c.Command.HasFlags() {
		resetArrayAndSliceFlags(c.Command)
	}
	c.Command.SetArgs(args)
	err := c.Command.Execute()
	errors.DisplaySuggestionsMessage(err, os.Stderr)
	analytics.SendAnalyticsAndLog(c.Command, args, err, c.analytics, c.logger)
	feedback.HandleFeedbackNudge(cliName, args)
	return err
}

func NewShellPrompt(rootCmd *cobra.Command, compl completer.Completer, cfg *v3.Config, logger *log.Logger, analytics analytics.Client, opts ...goprompt.Option) *ShellPrompt {
	shell := &ShellPrompt{
		Completer: compl,
		RootCmd: &instrumentedCommand{
			Command:   rootCmd,
			analytics: analytics,
			logger:    logger,
		},
		Config: cfg,
	}
	prompt := goprompt.New(promptExecutorFunc(cfg, shell), shell.Complete, opts...)
	shell.Prompt = prompt

	return shell
}

func promptExecutorFunc(cfg *v3.Config, shell *ShellPrompt) func(string) {
	return func(in string) {
		promptArgs, _ := shellparser.Fields(in, func(string) string { return "" })
		_ = shell.RootCmd.Execute(cfg.CLIName, promptArgs)
	}
}

func (p *ShellPrompt) Run() {
	p.RootCmd.InitDefaultHelpCmd()
	p.RootCmd.InitDefaultHelpFlag()
	p.Prompt.Run()
}

func DefaultPromptOptions() []goprompt.Option {
	return append([]goprompt.Option{
		goprompt.OptionShowCompletionAtStart(),
		goprompt.OptionPrefix(defaultShellPrefix),
		goprompt.OptionMaxSuggestion(maxSuggestion),
	}, DefaultColor256PromptOptions()...)
}

func DefaultColor256PromptOptions() []goprompt.Option {
	const powderSimilar = 195
	const denimSimilar = 17

	colorOpts := []goprompt.Option{
		goprompt.OptionPrefixTextColor(powderSimilar),
		goprompt.OptionPreviewSuggestionTextColor(powderSimilar),
		goprompt.OptionSuggestionBGColor(powderSimilar),
		goprompt.OptionSuggestionTextColor(denimSimilar),
		goprompt.OptionSelectedSuggestionBGColor(denimSimilar),
		goprompt.OptionSelectedSuggestionTextColor(powderSimilar),
		goprompt.OptionDescriptionBGColor(denimSimilar),
		goprompt.OptionDescriptionTextColor(powderSimilar),
		goprompt.OptionSelectedDescriptionBGColor(powderSimilar),
		goprompt.OptionSelectedDescriptionTextColor(denimSimilar),
		goprompt.OptionScrollbarBGColor(denimSimilar),
		goprompt.OptionScrollbarThumbColor(powderSimilar),
		goprompt.OptionInputTextColor(powderSimilar),
	}
	return append(colorOpts, goprompt.OptionWriter(NewStdoutColor256VT100Writer())) // Be mindful of order.
}
