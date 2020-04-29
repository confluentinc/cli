package completer

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CobraCompleter struct {
	RootCmd *cobra.Command
}

func NewCobraCompleter(rootCmd *cobra.Command) *CobraCompleter {
	return &CobraCompleter{
		RootCmd: rootCmd,
	}
}

func (c *CobraCompleter) Complete(d prompt.Document) []prompt.Suggest {
	matchedCmd := c.RootCmd
	args := strings.Fields(d.CurrentLine())
	var foundArgs []string
	if found, cmdArgs, err := matchedCmd.Find(args); err == nil {
		matchedCmd = found
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

	matchedCmd.LocalFlags().VisitAll(addFlags)
	matchedCmd.InheritedFlags().VisitAll(addFlags)

	if matchedCmd.HasAvailableSubCommands() {
		for _, cmd := range matchedCmd.Commands() {
			if !cmd.Hidden {
				suggestions = append(suggestions, prompt.Suggest{Text: cmd.Name(), Description: cmd.Short})
			}
		}
	}
	// Handle commands that are prefixes of each other (e.g. "a" and "aa")
	// This is necessary because (c *cobra.Command).Find() will only return the shortest matching command.
	if matchedCmd.HasParent() && matchedCmd.Parent().HasAvailableSubCommands() {
		for _, cmd := range matchedCmd.Parent().Commands() {
			if !cmd.Hidden {
				if strings.HasPrefix(cmd.Name(), matchedCmd.Name()) {
					suggestions = append(suggestions, prompt.Suggest{Text: cmd.Name(), Description: cmd.Short})
				}
			}
		}
	}
	if len(foundArgs) > 0 && matchedCmd.Args == nil {
		// Partial subcommand filter.
		if len(foundArgs) > 1 {
			// Can't have two partial subcommands.
			return []prompt.Suggest{}
		}
		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursorWithSpace(), true)
	}
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}
