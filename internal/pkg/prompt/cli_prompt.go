package prompt

import (
	"os"
	"strings"

	gprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/completer"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
)

const (
	defaultPrefix = " > "
)

type CLIPrompt struct {
	*gprompt.Prompt
	RootCmd   *cobra.Command
	Completer completer.Completer
	Config    *v2.Config
}

func NewCLIPrompt(rootCmd *cobra.Command, compl completer.Completer, cfg *v2.Config,
	validator pauth.TokenValidator, opts ...gprompt.Option) *CLIPrompt {
	cliPrompt := &CLIPrompt{
		Completer: compl,
		RootCmd:   rootCmd,
		Config:    cfg,
	}
	var extraOpts []gprompt.Option
	if masterCompleter, ok := compl.(*completer.MasterCompleter); ok {
		fuzzySearchOpt := gprompt.OptionAddKeyBind(
			gprompt.KeyBind{
				Key: gprompt.ControlF,
				Fn: func(buffer *gprompt.Buffer) {
					masterCompleter.FuzzyComplete = !masterCompleter.FuzzyComplete
				},
			},
		)
		extraOpts = append(extraOpts, fuzzySearchOpt)
	}
	livePrefixFunc := func() (prefix string, useLivePrefix bool) {
		hasLogin := cliPrompt.Config.HasLogin()
		var prefixEnding string
		if hasLogin {
			ctx := cliPrompt.Config.Context()
			token := ctx.State.AuthToken
			err := validator.ValidateToken(token)
			if err != nil {
				prefixEnding = "❌"
			} else {
				prefixEnding = "✅"
			}
		} else {
			prefixEnding = "❌"
		}
		prefix = cliPrompt.Config.CLIName + " " + prefixEnding
		if masterCompleter, ok := compl.(*completer.MasterCompleter); ok {
			if masterCompleter.FuzzyComplete {
				prefix += "  🔍"
			}
		}
		prefix += " > "
		return prefix, true
	}
	livePrefixOpt := gprompt.OptionLivePrefix(livePrefixFunc)
	extraOpts = append(extraOpts, livePrefixOpt)
	opts = append(opts, extraOpts...)
	prompt := gprompt.New(
		func(in string) {
			promptArgs := strings.Fields(in)
			os.Args = append([]string{os.Args[0]}, promptArgs...)
			rootCmd.Execute()
		},
		func(d gprompt.Document) []gprompt.Suggest {
			return cliPrompt.Completer.Complete(d)
		},
		opts...,
	)
	cliPrompt.Prompt = prompt
	return cliPrompt
}

func DefaultPromptOptions() []gprompt.Option {
	// Black is actually White and vice versa
	return []gprompt.Option{
		gprompt.OptionShowCompletionAtStart(),
		gprompt.OptionPrefixTextColor(gprompt.Blue),
		gprompt.OptionPreviewSuggestionTextColor(gprompt.Purple),
		gprompt.OptionSuggestionBGColor(gprompt.LightGray),
		gprompt.OptionSuggestionTextColor(gprompt.Black),
		gprompt.OptionSelectedSuggestionBGColor(gprompt.DarkBlue),
		gprompt.OptionSelectedSuggestionTextColor(gprompt.White),
		gprompt.OptionDescriptionBGColor(gprompt.DarkGray),
		gprompt.OptionDescriptionTextColor(gprompt.LightGray),
		gprompt.OptionSelectedDescriptionBGColor(gprompt.Blue),
		gprompt.OptionSelectedDescriptionTextColor(gprompt.White),
		gprompt.OptionScrollbarBGColor(gprompt.Blue),
		gprompt.OptionScrollbarThumbColor(gprompt.DarkBlue),
		gprompt.OptionPrefix(defaultPrefix),
		gprompt.OptionAddASCIICodeBind(
			[]gprompt.ASCIICodeBind{
				// Previous word (Alt-left arrow).
				{
					ASCIICode: []byte{0x1b, 0x62},
					Fn:        gprompt.GoLeftWord,
				},
				// Next word (Alt-right arrow).
				{
					ASCIICode: []byte{0x1b, 0x66},
					Fn:        gprompt.GoRightWord,
				},
			}...,
		),
	}
}
