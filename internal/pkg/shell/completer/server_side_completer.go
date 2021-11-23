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

type argType int

const (
	// non flag arguments
	argValue argType = iota
	// flag that still needs argument
	flagNeedArg
	// flag that doesn't expect an argument
	// either flag that takes no argument, e.g. -v, or flag that already has argument appended to it, e.g. --resource=lkc-123
	flagNoArg
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
		"confluent api-key create",
		"confluent api-key list",
		"confluent connect create",
		"confluent connect describe",
		"confluent connect list",
		"confluent connect plugin describe",
		"confluent connect plugin list",
		"confluent environment create",
		"confluent environment list",
		"confluent kafka acl create",
		"confluent kafka acl list",
		"confluent kafka cluster list",
		"confluent kafka cluster describe",
		"confluent kafka topic describe",
		"confluent kafka topic list",
		"confluent kafka region list",
		"confluent ksql app create",
		"confluent ksql app describe",
		"confluent ksql app list",
		"confluent price list",
		"confluent schema-registry cluster describe",
		"confluent schema-registry cluster enable",
		"confluent schema-registry schema create",
		"confluent schema-registry subject list",
		"confluent schema-registry subject describe",
		"confluent iam service-account list",
		"confluent iam service-account create",
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
	currentLine := d.CurrentLine()
	// must be typing a new argument to get suggestions
	// we will not support suggestions for flag when user types '=' e.g. --resource=
	// right now without space goprompt is going to replace the whole argument so it will override the flag name instead of appending
	// need to figure out a way around that before we can support '=' suggestion
	if !strings.HasSuffix(currentLine, " ") {
		return []prompt.Suggest{}
	}

	cmd := c.Root
	args := strings.Fields(currentLine)

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

	// List of argType corresponding to the arg position in the argument list
	// lastFlagWithArg is the flag name if last argument is flag that expects argument else ""
	argTypeList, lastFlagWithArg := c.getArgTypeListAndLastFlagWithArg(cmd, args)
	// If flagName is not empty then we are in a flag completion state
	if len(lastFlagWithArg) > 0 {
		suggestions := c.getSuggestionsForFlag(cmd, lastFlagWithArg)
		return filterSuggestions(d, suggestions)
	}

	if !c.canAcceptArgument(cmd, argTypeList) {
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

// Return list of argType where each element is the argument type of the corresponding arg in the arg list
// Also returns flag name if the last argument is a flag that expects argument, else just returns ""
func (c *ServerSideCompleterImpl) getArgTypeListAndLastFlagWithArg(cmd *cobra.Command, args []string) ([]argType, string) {
	lastFlagWithArgName := ""
	// initialized with all values as the default value of 0 which is argValue argType
	argTypeList := make([]argType, len(args))
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
			if utils.IsShorthandCountFlag(flag, arg) {
				argTypeList[i] = flagNoArg
				continue
			}
			candidate := arg
			if strings.Contains(arg, "=") {
				splitArgs := strings.Split(arg, "=")
				candidate = splitArgs[0]
			}
			if longName == candidate || shortName == candidate {
				if !utils.IsFlagWithArg(flag) {
					argTypeList[i] = flagNoArg
				} else if !strings.Contains(arg, "=") {
					argTypeList[i] = flagNeedArg
					if i == len(args)-1 {
						lastFlagWithArgName = flag.Name
					}
				} else if i == len(args)-1 && string(arg[len(arg)-1]) == "=" {
					argTypeList[i] = flagNeedArg
					lastFlagWithArgName = flag.Name

				} else {
					argTypeList[i] = flagNoArg
				}
			}

		}
	}

	cmd.LocalFlags().VisitAll(checkFlag)
	cmd.InheritedFlags().VisitAll(checkFlag)

	if lastFlagWithArgName != "" {
		if !c.checkFlagUse(argTypeList) {
			lastFlagWithArgName = ""
		}
	}
	return argTypeList, lastFlagWithArgName
}

// Check to prevent suggestions when the flag is passed as other flag values
// e.g. --resource --resource the suggestions should only show when the user types --resource the first time
// the second --resource would be a mistake and we do not want to show suggestions for that
func (c *ServerSideCompleterImpl) checkFlagUse(argTypeList []argType) bool {
	if len(argTypeList) == 0 {
		return false
	}
	i := 0
	for ; i < len(argTypeList); i += 1 {
		curArgType := argTypeList[i]
		if curArgType == flagNeedArg {
			if i == len(argTypeList)-1 {
				return true
			}
			i += 1
		}
	}
	return false
}

// Check if the command expects an argument
func (c *ServerSideCompleterImpl) canAcceptArgument(cmd *cobra.Command, argTypeList []argType) bool {
	argNum := c.argCount(argTypeList)
	tmpArgs := make([]string, argNum+1)
	err := cmd.ValidateArgs(tmpArgs)
	return err == nil
}

// Count the number of arguments for the command
// = (Number of arguments in total) - (Number of flags and flag values)
func (c *ServerSideCompleterImpl) argCount(argTypeList []argType) int {
	count := 0
	for i := 0; i < len(argTypeList); i++ {
		if argTypeList[i] == argValue {
			count += 1
		} else {
			if argTypeList[i] == flagNeedArg {
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
	// If it is the case that the cache is not yet updated, an empty suggestions is returned as querying now would hang for the user
	suggestions, _ = c.getCachedSuggestions(cc)
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
		// If not in cache just return emtpy list
		// It could either be because it is not in completable state or the cache udpate just didin't happen soon enough
		return []prompt.Suggest{}
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

func (c *ServerSideCompleterImpl) getCachedSuggestions(cc ServerCompletableCommand) ([]prompt.Suggest, bool) {
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
	var filtered []prompt.Suggest
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
