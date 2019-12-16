package cmd

import (
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CALLBACK_ANNOTATION
const CallbackAnnotation = "cobra-prompt"

// CobraPrompt requires RootCmd to run
type CobraPrompt struct {
	// RootCmd is the start point, all its sub commands and flags will be available as suggestions
	RootCmd *cobra.Command

	// GoPromptOptions is for customize go-prompt
	// see https://github.com/c-bata/go-prompt/blob/master/option.go
	GoPromptOptions []prompt.Option

	// DynamicSuggestionsFunc will be executed if an command has CALLBACK_ANNOTATION as an annotation. If it's included
	// the value will be provided to the DynamicSuggestionsFunc function.
	DynamicSuggestionsFunc func(annotation string, document prompt.Document) []prompt.Suggest

	// ResetFlagsFlag will add a new persistent flag to RootCmd. This flags can be used to turn off flags value reset
	ResetFlagsFlag bool
	FuzzyFind bool
	CLIName string
}

// Run will automatically generate suggestions for all cobra commands and flags defined by RootCmd
// and execute the selected commands. Run will also reset all given flags by default, see ResetFlagsFlag
func (co *CobraPrompt) Run() {
	co.prepare()
	p := prompt.New(
		func(in string) {
			promptArgs := strings.Fields(in)
			os.Args = append([]string{os.Args[0]}, promptArgs...)
			co.RootCmd.Execute()
		},
		func(d prompt.Document) []prompt.Suggest {
			return co.findSuggestions(d)
		},
		co.GoPromptOptions...,
	)
	p.Run()
}

func (co *CobraPrompt) prepare() {
	if co.ResetFlagsFlag {
		co.RootCmd.PersistentFlags().BoolP("flags-no-reset", "",
			false, "Flags will no longer reset to default value")
	}
}

func (co *CobraPrompt) findSuggestions(d prompt.Document) []prompt.Suggest {
	if co.FuzzyFind {
		matches := findMatchingCommands(d.CurrentLine(), co.RootCmd, []*cobra.Command{})
		var suggestions []prompt.Suggest
		for _, m := range matches {
			cmdPath := strings.TrimPrefix(m.CommandPath(), co.CLIName)
			suggestion := prompt.Suggest{
				Text:        cmdPath,
				Description: m.Short,
			}
			suggestions = append(suggestions, suggestion)
		}
		return suggestions
	}
	command := co.RootCmd
	args := strings.Fields(d.CurrentLine())
	var foundArgs []string
	if found, cmdArgs, err := command.Find(args); err == nil {
		command = found
		foundArgs = cmdArgs
	}
	var suggestions []prompt.Suggest
	resetFlags, _ := command.Flags().GetBool("flags-no-reset")
	addFlags := func(flag *pflag.Flag) {
		if flag.Changed && !resetFlags {
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

	annotation := command.Annotations[CallbackAnnotation]
	if co.DynamicSuggestionsFunc != nil && annotation != "" {
		suggestions = append(suggestions, co.DynamicSuggestionsFunc(annotation, d)...)
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

func findMatchingCommands(input string, rootCmd *cobra.Command, matches []*cobra.Command) []*cobra.Command {
	for _, command := range rootCmd.Commands() {
		if strings.Contains(command.Use, input) {
			matches = append(matches, command)
		}
		matches = findMatchingCommands(input, command, matches)
	}
	return matches
}
