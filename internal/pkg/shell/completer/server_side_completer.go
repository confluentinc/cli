package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ServerSideCompleterImpl struct {
	// map[string]ServerCompletableCommand
	commandsByPath *sync.Map
	// map[string][]prompt.Suggest
	cachedSuggestionsByPath *sync.Map

	Root *cobra.Command
}

func NewServerSideCompleter(root *cobra.Command) *ServerSideCompleterImpl {
	return &ServerSideCompleterImpl{
		Root:                    root,
		commandsByPath:          new(sync.Map),
		cachedSuggestionsByPath: new(sync.Map),
	}
}

// Complete
// if NOT in a completable state (spaces, not accepted, etc)
// 		RETURN
// if command is completable
// 		fetch and cache results
//		RETURN
// else if command is NOT a child of a completable command
// 		RETURN no results
// else
// 		if cached results are NOT available
// 			fetch and cache results
//		RETURN results
func (c *ServerSideCompleterImpl) Complete(d prompt.Document) []prompt.Suggest {
	cmd := c.Root
	args := strings.Fields(d.CurrentLine())

	if found, foundArgs, err := cmd.Find(args); err == nil {
		cmd = found
		args = foundArgs
	}

	if !c.inCompletableState(d, cmd, args) {
		return []prompt.Suggest{}
	}

	cc := c.getCompletableCommand(cmd)
	if cc != nil {
		go c.updateCachedSuggestions(cc)
		return []prompt.Suggest{}
	}

	if cc = c.getCompletableParent(cmd); cc == nil {
		return []prompt.Suggest{}
	}
	suggestions, ok := c.getCachedSuggestions(cc)
	if !ok {
		// Shouldn't happen, but just in case.
		// If this does happen then cache should be in the process of updating.
		suggestions = cc.ServerComplete()
	}
	return filterSuggestions(d, suggestions)
}

func (c *ServerSideCompleterImpl) updateCachedSuggestions(cc ServerCompletableCommand) {
	key := c.commandKey(cc.Cmd())
	c.cachedSuggestionsByPath.Store(key, cc.ServerComplete())
}

func (c *ServerSideCompleterImpl) getCachedSuggestions(cc ServerCompletableCommand) ([]prompt.Suggest, bool) {
	key := c.commandKey(cc.Cmd())
	v, ok := c.cachedSuggestionsByPath.Load(key)
	if !ok {
		return nil, false
	}
	return v.([]prompt.Suggest), true
}

// getCompletableCommand returns a matching ServerCompletableCommand, or nil if one is not found.
func (c *ServerSideCompleterImpl) getCompletableCommand(cmd *cobra.Command) ServerCompletableCommand {
	v, ok := c.commandsByPath.Load(c.commandKey(cmd))
	if !ok {
		return nil
	}
	return v.(ServerCompletableCommand)
}

// getCompletableParent return the completable parent if the specified command is a completable child,
// and false otherwise.
func (c *ServerSideCompleterImpl) getCompletableParent(cmd *cobra.Command) ServerCompletableCommand {
	parent := cmd.Parent()
	if parent == nil {
		return nil
	}
	cc := c.getCompletableCommand(parent)
	if cc == nil {
		return nil
	}
	for _, child := range cc.ServerCompletableChildren() {
		childKey := c.commandKey(child)
		matchedKey := c.commandKey(cmd)
		if childKey == matchedKey {
			return cc
		}
	}
	return nil
}

func filterSuggestions(d prompt.Document, suggestions []prompt.Suggest) []prompt.Suggest {
	filtered := []prompt.Suggest{}
	for _, suggestion := range suggestions {
		// only suggest if it does not appear anywhere in the input,
		// or if the suggestion is just a message to the user.
		// go-prompt filters out suggestions with empty string as text,
		// so we must suggest with at least one space.
		isMessage := strings.TrimSpace(suggestion.Text) == "" && suggestion.Description != ""
		if isMessage {
			// Introduce whitespace, or trim unnecessary whitespace.
			suggestion.Text = " "
		}
		if isMessage || !strings.Contains(d.Text, suggestion.Text) {
			filtered = append(filtered, suggestion)
		}
	}
	return filtered
}

func (c *ServerSideCompleterImpl) AddCommand(cmd ServerCompletableCommand) {
	c.commandsByPath.Store(c.commandKey(cmd.Cmd()), cmd)
}

func (c *ServerSideCompleterImpl) commandKey(cmd *cobra.Command) string {
	// trim CLI name
	return strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()+" ")
}

// inCompletableState checks whether the specified command is in a state where it should be considered for completion,
// which is:
// 1. when not after an uncompleted flag (api-key update --description)
// 2. when a command is not accepted (ending with a space)
// 3. when a command with a positional arg doesn't already have that arg provided
func (c *ServerSideCompleterImpl) inCompletableState(d prompt.Document, matchedCmd *cobra.Command, args []string) bool {
	var shouldSuggest = true

	// must be typing a new argument
	if !strings.HasSuffix(d.CurrentLine(), " ") {
		return false
	}

	// This is a heuristic to see if more args can be accepted. If no validation error occurs
	// for a number of args larger than the current number up to the chosen max, we say that more
	// args can be accepted. Cases where args only in some valid set (i.e: strings containing
	// the letter 'a') are accepted aren't considered for now.
	const maxReasonableArgs = 20
	canAcceptMoreArgs := false
	for i := len(args) + 1; i <= maxReasonableArgs; i++ {
		tmpArgs := make([]string, i)
		if err := matchedCmd.ValidateArgs(tmpArgs); err == nil {
			canAcceptMoreArgs = true
			break
		}
	}
	if !canAcceptMoreArgs {
		return false
	}

	_ = matchedCmd.ParseFlags(strings.Fields(d.CurrentLine()))

	addFlags := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
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
