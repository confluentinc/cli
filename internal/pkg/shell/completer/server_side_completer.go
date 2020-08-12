package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
	"sync"
)

type ServerSideCompleter struct {
	// map[string][]prompt.Suggest
	cachedSuggestionsByCmd *sync.Map

	// map[string]CompletionFunc
	suggestionFuncsByCmd *sync.Map

	// map[string]bool
	shouldSuggestForCmd *sync.Map

	RootCmd *cobra.Command
}

func NewServerSideCompleter(RootCmd *cobra.Command) *ServerSideCompleter {
	c := &ServerSideCompleter{
		cachedSuggestionsByCmd: new(sync.Map),
		suggestionFuncsByCmd:   new(sync.Map),
		shouldSuggestForCmd:	new(sync.Map),
		RootCmd:                RootCmd,
	}

	return c
}

func (c *ServerSideCompleter) Complete(d prompt.Document) []prompt.Suggest {

	cmd := c.RootCmd
	args := strings.Fields(d.CurrentLine())

	if found, _, err := cmd.Find(args); err == nil {
		cmd = found
	}

	// check if suggestion should occur
	if !c.shouldSuggestArgument(d, cmd) {
		return []prompt.Suggest{}
	}

	// check if child command with preloaded suggestions by the parent i.e api-key delete
	if cmd.Parent() != nil {
		parent := c.commandKey(cmd.Parent())
		if cachedSuggestions, ok := c.cachedSuggestionsByCmd.Load(parent); ok {
			if suggestions, ok := cachedSuggestions.([]prompt.Suggest); ok {
				filtered := []prompt.Suggest{}
				for _, suggestion := range suggestions {
					// only suggest if it does not appear anywhere in the input
					if !strings.Contains(d.Text, suggestion.Text) {
						filtered = append(filtered, suggestion)
					}
				}
				return filtered
			}
		}
	}

	// otherwise fetch the suggestions in the background if it is a parent command i.e api-key
	key := c.commandKey(cmd)
	if f, ok := c.suggestionFuncsByCmd.Load(key); ok {
		go c.updateSuggestion(c.commandKey(cmd), f.(func() []prompt.Suggest))
	}

	return []prompt.Suggest{}
}

func (c *ServerSideCompleter) AddCommand(cmd *cobra.Command, shouldSuggest bool) {
	c.shouldSuggestForCmd.Store(c.commandKey(cmd), shouldSuggest)
}

func (c *ServerSideCompleter) AddSuggestionFunction(cmd *cobra.Command, suggestionFunc func() []prompt.Suggest) {
	c.suggestionFuncsByCmd.Store(c.commandKey(cmd), suggestionFunc)
}

func (c *ServerSideCompleter) commandKey(cmd *cobra.Command) string {
	// trim CLI name
	return strings.TrimPrefix(cmd.CommandPath(), c.RootCmd.Name()+" ")
}

func (c *ServerSideCompleter) updateSuggestion(annotation string, suggestionFunc func() []prompt.Suggest) {
	suggestions := suggestionFunc()
	c.cachedSuggestionsByCmd.Store(annotation, suggestions)
}

// checks whether an argument should be suggested
// 1. when not after an uncompleted flag (api-key update --description)
// 2. when a command is not accepted (ending with a space)
// 3. when a command was not registered with a suggestion function
func (c *ServerSideCompleter) shouldSuggestArgument(d prompt.Document, matchedCmd *cobra.Command) bool {
	var shouldSuggest = true

	// must be typing a new argument
	if !strings.HasSuffix(d.CurrentLine(), " ") {
		return false
	}

	// must be registered for a suggestion function
	if registered, ok := c.shouldSuggestForCmd.Load(c.commandKey(matchedCmd)); ok && !registered.(bool)  {
		return false
	}

	matchedCmd.ParseFlags(strings.Fields(d.CurrentLine()))

	addFlags := func(flag *pflag.Flag) {
		if flag.Changed {
			flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		longName := "--" + flag.Name
		shortName := "-" + flag.Shorthand
		endsWithFlag := strings.HasSuffix(d.GetWordBeforeCursorWithSpace(), shortName+" ") ||
			strings.HasSuffix(d.GetWordBeforeCursorWithSpace(), longName+" ")
		if endsWithFlag {
			// should not suggest an argument if flag is not completed with a value but expects one
			if flag.DefValue == "" || flag.DefValue == "0" {
				shouldSuggest = false
			}
		}
	}

	matchedCmd.LocalFlags().VisitAll(addFlags)
	matchedCmd.InheritedFlags().VisitAll(addFlags)
	return shouldSuggest
}
