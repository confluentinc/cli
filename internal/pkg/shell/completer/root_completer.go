package completer

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RootCompleter struct {
	RootCmd *cobra.Command
}

func NewRootCompleter(rootCmd *cobra.Command) *RootCompleter {
	return &RootCompleter{
		RootCmd: rootCmd,
	}
}

func (c *RootCompleter) Complete(d prompt.Document) []prompt.Suggest {
	command := c.RootCmd
	args := strings.Fields(d.CurrentLine())
	var foundArgs []string
	if found, cmdArgs, err := command.Find(args); err == nil {
		command = found
		foundArgs = cmdArgs
	}

	var suggestions []prompt.Suggest
	addFlags := func(flag *pflag.Flag) {
		if flag.Changed {
			flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		if strings.HasPrefix(d.GetWordBeforeCursor(), "--") {
			suggestions = append(suggestions, prompt.Suggest{Text: "--" + flag.Name, Description: flag.Usage})
		} else if strings.HasPrefix(d.GetWordBeforeCursor(), "-") && flag.Shorthand != "" {
			suggestions = append(suggestions, prompt.Suggest{Text: "-" + flag.Shorthand, Description: flag.Usage})
		}
	}

	command.LocalFlags().VisitAll(addFlags)
	command.InheritedFlags().VisitAll(addFlags)

	if command.HasAvailableSubCommands() {
		for _, c := range command.Commands() {
			if !c.Hidden {
				suggestions = append(suggestions, prompt.Suggest{Text: c.Name(), Description: c.Short})
			}
		}
	}

	if len(foundArgs) > 0 {
		for _, suggestion := range suggestions {
			if foundArgs[0] == suggestion.Text {
				return []prompt.Suggest{}
			}
		}
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}
