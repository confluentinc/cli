package completer

import (
	"fmt"
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
	line := d.CurrentLine()
	suggestionAccepted := strings.HasSuffix(line, " ")
	args := strings.Fields(line)
	filter := ""
	if !suggestionAccepted && len(args) > 0 {
		filter = args[len(args)-1]
		args = args[:len(args)-1]
	}
	var foundArgs []string
	found, cmdArgs, err := matchedCmd.Find(args)
	var suggestions []prompt.Suggest
	if err == nil {
		matchedCmd = found
	}
	foundArgs = cmdArgs

	suggestions = append(suggestions, getFlagSuggestions(d, matchedCmd)...)
	// TODO: Check if persistent flags are handled.

	if matchedCmd.HasAvailableSubCommands() {
		for _, cmd := range matchedCmd.Commands() {
			if !cmd.Hidden {
				suggestions = addCmdToSuggestions(suggestions, cmd)
			}
		}
	}
	_ = matchedCmd.ParseFlags(args)
	pathWithoutRoot := strings.TrimPrefix(matchedCmd.CommandPath(), c.RootCmd.Name())
	pathWithoutRoot = strings.TrimSpace(pathWithoutRoot)
	allArgs := matchedCmd.Flags().Args()
	unmatchedArgs := strings.TrimPrefix(strings.Join(allArgs, " "), pathWithoutRoot)
	unmatchedArgArr := strings.Fields(unmatchedArgs)
	if len(unmatchedArgArr) == 0 {
		return prompt.FilterHasPrefix(suggestions, filter, true)
	}
	// Filter is all args + flags starting from first unmatched arg.
	// Find index first unmatched arg.
	if unmatchedArgArr[0] == matchedCmd.Name() {
		unmatchedArgs = unmatchedArgs[1:]
	}

	unmatchedIndex := firstOccurrence(foundArgs, unmatchedArgArr[0])
	if unmatchedIndex == -1 {
		panic(fmt.Sprint(foundArgs, unmatchedArgs))
	}
	filterSuffix := filter
	filter = strings.Join(foundArgs[unmatchedIndex:], " ")
	filter = strings.TrimPrefix(filter, pathWithoutRoot)

	filter += " " + filterSuffix
	if len(strings.TrimSpace(filter)) == 0 {
		filter = ""
	}
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
			flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		if strings.HasPrefix(d.GetWordBeforeCursorWithSpace(), "--") {
			suggestions = append(suggestions, prompt.Suggest{Text: "--" + flag.Name, Description: flag.Usage})
		} else if strings.HasPrefix(d.GetWordBeforeCursorWithSpace(), "-") && flag.Shorthand != "" {
			suggestions = append(suggestions, prompt.Suggest{Text: "-" + flag.Shorthand, Description: flag.Usage})
		}
	}

	matchedCmd.LocalFlags().VisitAll(addFlags)
	matchedCmd.InheritedFlags().VisitAll(addFlags)
	return suggestions
}
