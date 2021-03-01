package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type ServerSideCompleterImpl struct {
	// map[string]interface{} (value must be either ServerCompletableCommand or ServerCompletableFlag or both)
	commandsByPath *sync.Map
	// map[string][]prompt.Suggest
	cachedSuggestionsByPath *sync.Map

	// map[string][]prompt.Suggest
	staticFlagSuggestions *sync.Map
	// map[string]map[string]bool: key=flag value=set of command paths
	staticFlagSuggestionsCommandsMap *sync.Map

	Root *cobra.Command
}

func NewServerSideCompleter(root *cobra.Command) *ServerSideCompleterImpl {
	serverSideCompleterImpl := &ServerSideCompleterImpl{
		Root:                    root,
		commandsByPath:          new(sync.Map),
		cachedSuggestionsByPath: new(sync.Map),
	}
	serverSideCompleterImpl.initializeStaticFlagSuggestions()
	return serverSideCompleterImpl
}

func (c *ServerSideCompleterImpl) initializeStaticFlagSuggestions() {
	c.staticFlagSuggestions = new(sync.Map)
	c.staticFlagSuggestionsCommandsMap = new(sync.Map)
	outputFlagName := "output"
	outputSuggestions := []prompt.Suggest{
		{
			Text:        "human",
			Description: "human",
		},
		{
			Text:        "json",
			Description: "JSON",
		},
		{
			Text:        "yaml",
			Description: "YAML",
		},
	}
	outputCommandPaths := []string{
		"ccloud api-key create",
		"ccloud api-key list",
		"ccloud connector create",
		"ccloud connector describe",
		"ccloud connector list",
		"ccloud connector-catalog describe",
		"ccloud connector-catalog list",
		"ccloud environment create",
		"ccloud environment list",
		"ccloud kafka acl create",
		"ccloud kafka acl list",
		"ccloud kafka cluster list",
		"ccloud kafka cluster describe",
		"ccloud kafka topic describe",
		"ccloud kafka topic list",
		"ccloud kafka region list",
		"ccloud ksql app create",
		"ccloud ksql app describe",
		"ccloud ksql app list",
		"ccloud price list",
		"ccloud schema-registry cluster describe",
		"ccloud schema-registry cluster enable",
		"ccloud schema-registry schema create",
		"ccloud schema-registry subject list",
		"ccloud schema-registry subject describe",
		"ccloud service-account list",
		"ccloud service-account create",
	}
	c.AddStaticFlagCompletion(outputFlagName, outputSuggestions, outputCommandPaths)
}

func (c *ServerSideCompleterImpl) AddStaticFlagCompletion(flagName string, suggestions []prompt.Suggest, commandPaths []string) {
	c.staticFlagSuggestions.Store(flagName, suggestions)
	commandSet := make(map[string]bool)
	for _, commandPath := range commandPaths {
		commandSet[commandPath] = true
	}
	c.staticFlagSuggestionsCommandsMap.Store(flagName, commandSet)
}

func (c *ServerSideCompleterImpl) Complete(d prompt.Document) []prompt.Suggest {
	// must be typing a new argument to get suggestions
	if !strings.HasSuffix(d.CurrentLine(), " ") {
		return []prompt.Suggest{}
	}

	cmd := c.Root
	args := strings.Fields(d.CurrentLine())

	if found, foundArgs, err := cmd.Find(args); err == nil {
		cmd = found
		args = foundArgs
	}

	// If command is in commandsByPath map then it is a parent of commands with completion
	// Update the cache in the background
	// Return empty suggestions as the children of this command are the ones that need suggestions
	v, ok := c.commandsByPath.Load(c.commandKey(cmd))
	if ok {
		go c.updateCachedSuggestions(cmd, v)
		return []prompt.Suggest{}
	}

	// List where elements correspondes to the arg list, if the arg is a flag then flag in that position else nil
	flagList := c.getFlagList(cmd, args)
	// If flagName is not empty then we are in a flag completion state
	flagName := c.getFlagWithArg(flagList)
	if len(flagName) > 0 {
		suggestions := c.getSuggestionsForFlag(cmd, flagName)
		return filterSuggestions(d, suggestions)
	}

	if !c.canAcceptArgument(cmd, flagList) {
		return []prompt.Suggest{}
	}
	suggestions := c.getSuggestionsForCommand(d, cmd)
	return filterSuggestions(d, suggestions)
}

// Store suggestions in cache
// Cache is updated for the children commands of the command the user is currently typing
// e.g. ccloud api-key -> updates cache for ccloud api-key delete, or --resource flags for the various ccloud api-key command
// If command cannot be completed, e.g. user not logged in, then add empty list to reset the cache
func (c *ServerSideCompleterImpl) updateCachedSuggestions(cmd *cobra.Command, v interface{}) {
	canComplete := pcmd.CanCompleteCommand(cmd)
	cc, _ := v.(ServerCompletableCommand)
	cf, _ := v.(ServerCompletableFlag)
	if cc != nil {
		cmd := cc.Cmd()
		key := c.commandKey(cmd)
		var suggestions []prompt.Suggest
		if canComplete {
			suggestions = cc.ServerComplete()
		}
		c.cachedSuggestionsByPath.Store(key, suggestions)
	}

	if cf != nil {
		for flagName, completeFunc := range cf.ServerFlagComplete() {
			var suggestions []prompt.Suggest
			if canComplete {
				suggestions = completeFunc()
			}
			c.cachedSuggestionsByPath.Store(c.flagKey(cf.Cmd(), flagName), suggestions)
		}
	}
}

// Return list of the same length as the args list
// Each element in the list corresponds to the arg of the same position
// If the arg is a flag then the value is *pflag.Flag, else nil
func (c *ServerSideCompleterImpl) getFlagList(cmd *cobra.Command, args []string) []*pflag.Flag {
	flagList := make([]*pflag.Flag, len(args))
	checkFlag := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		longName := "--" + flag.Name
		shortName := "-" + flag.Shorthand
		for i, arg := range args {
			if utils.IsShorthandCountFlag(flag, arg) || longName == arg || shortName == arg {
				flagList[i] = flag
			}
		}
	}

	cmd.LocalFlags().VisitAll(checkFlag)
	cmd.InheritedFlags().VisitAll(checkFlag)
	return flagList
}

// Return flag name if the last argument is a flag that accepts argument, else return ""
func (c *ServerSideCompleterImpl) getFlagWithArg(flagList []*pflag.Flag) string {
	if len(flagList) == 0 {
		return ""
	}
	lastIndex := len(flagList) - 1
	lastFlag := flagList[lastIndex]
	if lastFlag == nil || !utils.IsFlagWithArg(lastFlag) {
		return ""
	}
	// Check to prevent suggestions when the flag is passed as other flag values
	// e.g. --resource --resource the suggestions should only show when the user types --resource the first time
	// the second --resource would be a mistake and we do not want to show suggestions for that
	i := 0
	for ; i <= lastIndex; i += 1 {
		curFlag := flagList[i]
		if i == lastIndex && utils.IsFlagWithArg(curFlag) {
			return curFlag.Name
		}
		if utils.IsFlagWithArg(flagList[i]) {
			i += 1
		}
	}
	return ""
}

// Check if the command expects an argument
func (c *ServerSideCompleterImpl) canAcceptArgument(cmd *cobra.Command, flagList []*pflag.Flag) bool {
	argNum := c.argCount(flagList)
	tmpArgs := make([]string, argNum+1)
	err := cmd.ValidateArgs(tmpArgs)
	return err == nil
}

// Count the number of arguments for the command
// = (Number of arguments in total) - (Number of flags and flag values)
func (c *ServerSideCompleterImpl) argCount(flagList []*pflag.Flag) int {
	count := 0
	for i := 0; i < len(flagList); i++ {
		if flagList[i] == nil {
			count += 1
		} else {
			if utils.IsFlagWithArg(flagList[i]) {
				i += 1
			}
		}
	}
	return count
}

// Check the cache for suggestions
func (c *ServerSideCompleterImpl) getSuggestionsForCommand(d prompt.Document, cmd *cobra.Command) []prompt.Suggest {
	var suggestions []prompt.Suggest
	var cc ServerCompletableCommand
	// Find the parent command that made the queries to update the cache
	if cc = c.getCompletableParent(cmd); cc == nil {
		return suggestions
	}
	// If found parent then we expect suggestions in the cache
	// If not found in cache it is either not in a completable state or the cache is not yet updated which shouldn't really happen
	suggestions, ok := c.getCachedSuggestions(d, cc)
	if !ok {
		if !pcmd.CanCompleteCommand(cmd) {
			return suggestions
		}
		// Call the ServerComplete func directly in case the cache didn't update in time
		suggestions = cc.ServerComplete()
	}
	return suggestions
}

// Check for flag suggestions, return empty list if not found or not in the state for flag suggestions
func (c *ServerSideCompleterImpl) getSuggestionsForFlag(cmd *cobra.Command, flagName string) []prompt.Suggest {
	// check static flag
	v, ok := c.staticFlagSuggestionsCommandsMap.Load(flagName)
	if ok {
		commandSet := v.(map[string]bool)
		if _, ok = commandSet[cmd.CommandPath()]; ok {
			v, ok := c.staticFlagSuggestions.Load(flagName)
			if !ok {
				return []prompt.Suggest{}
			}
			return v.([]prompt.Suggest)
		}
	}

	parent := c.getParentServerCompletableFlag(cmd, flagName)
	if parent == nil {
		return []prompt.Suggest{}
	}

	v, ok = c.cachedSuggestionsByPath.Load(c.flagKey(parent.Cmd(), flagName))
	if !ok {
		// in case caching didn't finish
		return parent.ServerFlagComplete()[flagName]()
	}
	return v.([]prompt.Suggest)
}

// Check that parent of current command is ServerCompletableFlag
// and that the current command is the completable child for the specific flag
func (c *ServerSideCompleterImpl) getParentServerCompletableFlag(cmd *cobra.Command, flagName string) ServerCompletableFlag {
	// check parent
	parent := cmd.Parent()
	if parent == nil {
		return nil
	}
	v, ok := c.commandsByPath.Load(c.commandKey(parent))
	if !ok {
		return nil
	}

	cf, ok := v.(ServerCompletableFlag)
	if !ok {
		return nil
	}

	childCommads, ok := cf.ServerCompletableFlagChildren()[flagName]
	if !ok {
		return nil
	}

	for _, childCmd := range childCommads {
		if childCmd.CommandPath() == cmd.CommandPath() {
			return cf
		}
	}

	return nil
}

func (c *ServerSideCompleterImpl) getCachedSuggestions(d prompt.Document, cc ServerCompletableCommand) ([]prompt.Suggest, bool) {
	key := c.commandKey(cc.Cmd())
	v, ok := c.cachedSuggestionsByPath.Load(key)
	if !ok {
		return nil, false
	}
	return v.([]prompt.Suggest), true
}

// Checks that the parent of the current command is a ServerCompletableCommand
// and that the current command is the child of that ServerCompletableCommand
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

// Returns a matching ServerCompletableCommand, or nil if one is not found.
func (c *ServerSideCompleterImpl) getCompletableCommand(cmd *cobra.Command) ServerCompletableCommand {
	v, ok := c.commandsByPath.Load(c.commandKey(cmd))
	if !ok {
		return nil
	}
	return v.(ServerCompletableCommand)
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

func (c *ServerSideCompleterImpl) AddCommand(cmd interface{}) {
	cc, ok := cmd.(ServerCompletableCommand)
	if ok {
		c.commandsByPath.Store(c.commandKey(cc.Cmd()), cc)
		return
	}
	cf, ok := cmd.(ServerCompletableFlag)
	if ok {
		c.commandsByPath.Store(c.commandKey(cf.Cmd()), cf)
		return
	}
	panic("Command added must implement either ServerCompletableCommand or ServerCompletableFlag or both.")
}

func (c *ServerSideCompleterImpl) commandKey(cmd *cobra.Command) string {
	return strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()+" ")
}

func (c *ServerSideCompleterImpl) flagKey(cmd *cobra.Command, flagName string) string {
	commandName := strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()+" ")
	return commandName + " --" + flagName
}
