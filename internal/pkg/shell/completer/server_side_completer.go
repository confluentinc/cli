package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ServerSideCompleterImpl struct {
	// map[string]*cachingCommand
	commandsByPath *sync.Map

	Root *cobra.Command
}

func NewServerSideCompleter(root *cobra.Command) *ServerSideCompleterImpl {
	c := &ServerSideCompleterImpl{
		Root:           root,
		commandsByPath: new(sync.Map),
	}

	return c
}

type cachingCommand struct {
	ServerCompletableCommand
	cachedSuggestions []prompt.Suggest
	mux               sync.Mutex
}

func (c *cachingCommand) ServerComplete() []prompt.Suggest {
	if c.cachedSuggestions != nil {
		c.refreshCache()
	}
	return c.cachedSuggestions
}

func (c *cachingCommand) refreshCache() {
	//c.mux.Lock()
	//defer c.mux.Unlock()
	c.cachedSuggestions = c.ServerCompletableCommand.ServerComplete()
}

// Complete 
// high level
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

	if found, _, err := cmd.Find(args); err == nil {
		cmd = found
	}

	if !c.inCompletableState(d, cmd) {
		return []prompt.Suggest{}
	}

	cc := c.getCompletableCommand(cmd)
	if cc != nil {
		go cc.refreshCache()
		return []prompt.Suggest{}
	}

	if cc = c.getCompletableParent(cmd); cc == nil {
		return []prompt.Suggest{}
	}

	return cc.ServerComplete()
}

// getCompletableCommand returns a matching ServerCompletableCommand, or nil if one is not found.
func (c *ServerSideCompleterImpl) getCompletableCommand(cmd *cobra.Command) *cachingCommand {
	v, ok := c.commandsByPath.Load(c.commandKey(cmd))
	if !ok {
		return nil
	}
	return v.(*cachingCommand)
}

// getCompletableParent return the completable parent if the specified command is a completable child,
// and false otherwise.
func (c *ServerSideCompleterImpl) getCompletableParent(cmd *cobra.Command) *cachingCommand {
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
		// only suggest if it does not appear anywhere in the input
		if !strings.Contains(d.Text, suggestion.Text) {
			filtered = append(filtered, suggestion)
		}
	}
	return filtered
}

func (c *ServerSideCompleterImpl) AddCommand(cmd ServerCompletableCommand) {
	cc := &cachingCommand{
		ServerCompletableCommand: cmd,
		cachedSuggestions:        []prompt.Suggest{},
		mux:                      sync.Mutex{},
	}
	c.commandsByPath.Store(c.commandKey(cmd.Cmd()), cc)
}

func (c *ServerSideCompleterImpl) commandKey(cmd *cobra.Command) string {
	// trim CLI name
	return strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()+" ")
}

// TODO: Implement
func (c *ServerSideCompleterImpl) updateSuggestion(cmd ServerCompletableCommand) {
}

// inCompletableState checks whether the specified command is in a state where it should be considered for completion,
// which is determined by the following:
// 1. when not after an uncompleted flag (api-key update --description)
// 2. when a command is not accepted (ending with a space)
func (c *ServerSideCompleterImpl) inCompletableState(d prompt.Document, matchedCmd *cobra.Command) bool {
	var shouldSuggest = true

	// must be typing a new argument
	if !strings.HasSuffix(d.CurrentLine(), " ") {
		return false
	}

	//// must be a completable child.
	//if matchedCmd.Parent() == nil {
	//	return false
	//}
	//parent := c.commandKey(matchedCmd.Parent())
	//v, ok := c.commandsByPath.Load(parent)
	//if !ok {
	//	// Should not happen.
	//	return false
	//}
	//cc := v.(ServerCompletableCommand)
	//hasCompletableChild := false
	//for _, child := range cc.ServerCompletableChildren() {
	//	childKey := c.commandKey(child)
	//	matchedKey := c.commandKey(matchedCmd)
	//	hasCompletableChild = childKey == matchedKey
	//	if hasCompletableChild {
	//		break
	//	}
	//}
	//if !hasCompletableChild {
	//	return false
	//}

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
