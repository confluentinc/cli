package completer

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type FuzzyCompleter struct {
	RootCmd *cobra.Command
	CLIName string
}

func NewFuzzyCompleter(rootCmd *cobra.Command, cliName string) *FuzzyCompleter {
	return &FuzzyCompleter{
		RootCmd: rootCmd,
		CLIName: cliName,
	}
}

func (c *FuzzyCompleter) Complete(d prompt.Document) []prompt.Suggest {
	matches := findMatchingCommands(d.CurrentLine(), c.RootCmd, []*cobra.Command{})
	var suggestions []prompt.Suggest
	for _, m := range matches {
		cmdPath := strings.TrimPrefix(m.CommandPath(), c.CLIName)
		suggestion := prompt.Suggest{
			Text:        cmdPath,
			Description: m.Short,
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions
}

func findMatchingCommands(input string, rootCmd *cobra.Command, matches []*cobra.Command) []*cobra.Command {
	for _, command := range rootCmd.Commands() {
		if strings.Contains(command.Use, input) {
			matches = append(matches, command)
		}
		matches = findMatchingCommands(input, command, matches)
	}
	return matches
}
