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
	filter := ""
	args := strings.Fields(d.CurrentLine())

	suggestionAccepted := strings.HasSuffix(d.CurrentLine(), " ")
	if !suggestionAccepted && len(args) > 0 {
		filter = args[len(args)-1]
		args = args[:len(args)-1]
	}

	matchedCmd, foundArgs, err := matchedCmd.Find(args)
	if err != nil {
		return []prompt.Suggest{}
	}

	suggestions := append([]prompt.Suggest{}, getFlagSuggestions(d, matchedCmd)...)
	if len(suggestions) > 0 {
		// if flags are suggested, don't need command suggestions
		return suggestions
	}

	if matchedCmd.HasAvailableSubCommands() {
		for _, cmd := range matchedCmd.Commands() {
			if !cmd.Hidden {
				suggestions = addCmdToSuggestions(suggestions, cmd)
			}
		}
	}

	_ = matchedCmd.ParseFlags(args)

	// handle completed flags
	matchedCmd.Flags().VisitAll(func(f *pflag.Flag) {
		for i, arg := range foundArgs {
			if "--"+f.Name == arg || "-"+f.Shorthand == arg {
				foundArgs = append(foundArgs[:i], foundArgs[i+1:]...)
				break
			}
		}
	})

	allArgs := matchedCmd.Flags().Args()
	pathWithoutRoot := strings.TrimPrefix(matchedCmd.CommandPath(), c.RootCmd.Name()+" ")
	unmatchedArgs := strings.TrimPrefix(strings.Join(allArgs, " "), pathWithoutRoot)
	unmatchedArgArr := append(strings.Fields(unmatchedArgs), foundArgs...)
	if len(unmatchedArgArr) == 0 {
		return prompt.FilterHasPrefix(suggestions, filter, true)
	}

	// Filter is all args + flags starting from first unmatched arg.
	unmatchedIndex := firstOccurrence(foundArgs, unmatchedArgArr[0])
	filterSuffix := filter
	filter = strings.Join(foundArgs[unmatchedIndex:], " ")
	filter = strings.TrimPrefix(filter, pathWithoutRoot)
	filter += " " + filterSuffix

	return prompt.FilterHasPrefix(suggestions, filter, true)
}

func firstOccurrence(list []string, str string) int {
	for i, s := range list {
		if s == str {
			return i
		}
	}
	return -1
}

func addCmdToSuggestions(suggestions []prompt.Suggest, cmd *cobra.Command) []prompt.Suggest {
	return append(suggestions, prompt.Suggest{Text: cmd.Name(), Description: cmd.Short})
}

func getFlagSuggestions(d prompt.Document, matchedCmd *cobra.Command) []prompt.Suggest {
	var suggestions []prompt.Suggest
	addFlags := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		longName := "--" + flag.Name
		shortName := "-" + flag.Shorthand
		flagUsed := strings.Contains(d.CurrentLine(), shortName+" ") || strings.Contains(d.CurrentLine(), longName+" ")
		if !flagUsed {
			if strings.HasPrefix(d.GetWordBeforeCursor(), "--") {
				suggestions = append(suggestions, prompt.Suggest{Text: longName, Description: flag.Usage})
			} else if strings.HasPrefix(d.GetWordBeforeCursor(), "-") && flag.Shorthand != "" {
				suggestions = append(suggestions, prompt.Suggest{Text: shortName, Description: flag.Usage})
			}
		}
	}

	matchedCmd.LocalFlags().VisitAll(addFlags)
	matchedCmd.InheritedFlags().VisitAll(addFlags)
	return suggestions
}
