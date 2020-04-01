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
	writer := NewStdoutTrueColorWriter()
	extraOpts = append(extraOpts, livePrefixOpt, gprompt.OptionWriter(writer))
	opts = append(extraOpts, opts...)
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

func DefaultTrueColorPromptOptions() []gprompt.Option {
	// Black is actually White and vice versa
	const island = 0x38CCED
	const academy = 0x0074A2
	const powder = 0xD7EFF6
	const denim = 0x173361
	const sky = 0x81CFE2
	const robinsEgg = 0xB4E1E4
	const sunrise = 0xF26135
	const ice = 0xE6F5FB

	colorOpts := []gprompt.Option{
		gprompt.OptionPrefixTextColor(powder),
		gprompt.OptionPreviewSuggestionTextColor(powder),
		gprompt.OptionSuggestionBGColor(powder),
		gprompt.OptionSuggestionTextColor(denim),
		gprompt.OptionSelectedSuggestionBGColor(denim),
		gprompt.OptionSelectedSuggestionTextColor(powder),
		gprompt.OptionDescriptionBGColor(denim),
		gprompt.OptionDescriptionTextColor(powder),
		gprompt.OptionSelectedDescriptionBGColor(powder),
		gprompt.OptionSelectedDescriptionTextColor(denim),
		gprompt.OptionScrollbarBGColor(denim),
		gprompt.OptionScrollbarThumbColor(powder),
	}
	writer := NewStdoutTrueColorWriter()
	colorOpts = append(colorOpts, gprompt.OptionWriter(writer))
	return append(colorOpts, defaultPromptOptions()...)
}

func DefaultColor256PromptOptions() []gprompt.Option {
	const powderSimilar = 152
	const denimSimilar = 17

	writer := NewStdoutColor256Writer()
	colorOpts := []gprompt.Option{
		gprompt.OptionPrefixTextColor(powderSimilar),
		gprompt.OptionPreviewSuggestionTextColor(powderSimilar),
		gprompt.OptionSuggestionBGColor(powderSimilar),
		gprompt.OptionSuggestionTextColor(denimSimilar),
		gprompt.OptionSelectedSuggestionBGColor(denimSimilar),
		gprompt.OptionSelectedSuggestionTextColor(powderSimilar),
		gprompt.OptionDescriptionBGColor(denimSimilar),
		gprompt.OptionDescriptionTextColor(powderSimilar),
		gprompt.OptionSelectedDescriptionBGColor(powderSimilar),
		gprompt.OptionSelectedDescriptionTextColor(denimSimilar),
		gprompt.OptionScrollbarBGColor(denimSimilar),
		gprompt.OptionScrollbarThumbColor(powderSimilar),
	}
	colorOpts = append(colorOpts, defaultPromptOptions()...)
	return append(colorOpts, gprompt.OptionWriter(writer)) // Be mindful of order.
}

func defaultPromptOptions() []gprompt.Option {
	return []gprompt.Option{
		gprompt.OptionShowCompletionAtStart(),
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
