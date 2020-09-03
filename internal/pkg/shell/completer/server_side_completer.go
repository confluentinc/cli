package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ServerSideCompleter struct {
	// map[string]CompletableCommand
	commandsByPath *sync.Map

	RootCmd *cobra.Command
}

func NewServerSideCompleter(RootCmd *cobra.Command) *ServerSideCompleter {
	c := &ServerSideCompleter{
		commandsByPath: new(sync.Map),
		RootCmd:        RootCmd,
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
	// if parent, load suggestions (in background).
	// if child, suggest.
	if cmd.Parent() != nil {
		parent := c.commandKey(cmd.Parent())
		v, ok := c.commandsByPath.Load(parent)
		if !ok {
			// Should not happen.
			return []prompt.Suggest{}
		}
		cc := v.(CompletableCommand)
		return cc.Complete()
		//if cachedSuggestions, ok := c.cachedSuggestionsByCmd.Load(parent); ok {
		//	if suggestions, ok := cachedSuggestions.([]prompt.Suggest); ok {
		//		return filterSuggestions(d, suggestions)
		//	}
		//}
	}

	// otherwise fetch the suggestions in the background if it is a parent command i.e api-key
	//key := c.commandKey(cmd)
	//if f, ok := c.suggestionFuncsByCmd.Load(key); ok {
	//	go c.updateSuggestion(cmd)
	//}

	return []prompt.Suggest{}
}

func filterSuggestions(d prompt.Document, suggestions []prompt.Suggest) []prompt.Suggest {
	filtered := []prompt.Suggest{}
	for _, suggestion := range suggestions {
		// only suggest if it does not appear anywhere in the input
		if !strings.Contains(d.Text, suggestion.Text) {
			filtered = append(filtered, suggestion)
		}
	}
	return filtered
}

func (c *ServerSideCompleter) AddCommand(cmd CompletableCommand) {
	c.commandsByPath.Store(c.commandKey(cmd.Cmd()), cmd)
}

func (c *ServerSideCompleter) commandKey(cmd *cobra.Command) string {
	// trim CLI name
	return strings.TrimPrefix(cmd.CommandPath(), c.RootCmd.Name()+" ")
}

// TODO: Implement
func (c *ServerSideCompleter) updateSuggestion(cmd CompletableCommand) {
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

	// must be a completable child.
	if matchedCmd.Parent() == nil {
		return false
	}
	parent := c.commandKey(matchedCmd.Parent())
	v, ok := c.commandsByPath.Load(parent)
	if !ok {
		// Should not happen.
		return false
	}
	cc := v.(CompletableCommand)
	hasCompletableChild := false
	for _, child := range cc.CompletableChildren() {
		childKey := c.commandKey(child)
		matchedKey := c.commandKey(matchedCmd)
		hasCompletableChild = childKey == matchedKey
		if hasCompletableChild {
			break
		}
	}
	if !hasCompletableChild {
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
