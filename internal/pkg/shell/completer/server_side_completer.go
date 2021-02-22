package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	shellparser "mvdan.cc/sh/v3/shell"
)

type ServerSideCompleterImpl struct {
	// map[string]ServerCompletableCommand
	commandsByPath *sync.Map
	// map[string]ServerCompletableFlag
	flagsByPath *sync.Map
	// map[string][]prompt.Suggest
	cachedSuggestionsByPath *sync.Map

	// map[string][]prompt.Suggest
	staticFlagSuggestions *sync.Map

	// map[string]map[string]bool: key=flag value=command path set
	staticFlagSuggestionsCommandMap *sync.Map

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
	c.staticFlagSuggestions.Store("output", []prompt.Suggest{
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
	})
	c.staticFlagSuggestionsCommandMap = new(sync.Map)
	c.staticFlagSuggestionsCommandMap.Store("output", map[string]bool{
		"ccloud api-key create":                   true,
		"ccloud api-key list":                     true,
		"ccloud connector create":                 true,
		"ccloud connector describe":               true,
		"ccloud connector list":                   true,
		"ccloud connector-catalog describe":       true,
		"ccloud connector-catalog list":           true,
		"ccloud environment list":                 true,
		"ccloud environment create":               true,
		"ccloud kafka cluster list":               true,
		"ccloud kafka topic describe":             true,
		"ccloud kafka acl create":                 true,
		"ccloud kafka acl list":                   true,
		"ccloud kafka region list":                true,
		"ccloud ksql app create":                  true,
		"ccloud ksql app describe":                true,
		"ccloud ksql app list":                    true,
		"ccloud price list":                       true,
		"ccloud schema-registry cluster describe": true,
		"ccloud schema-registry cluster enable":   true,
		"ccloud schema-registry schema create":    true,
		"ccloud schema-registry subject list":     true,
		"ccloud schema-registry subject describe": true,
		"ccloud service-account list":             true,
		"ccloud service-account create":           true,
	})
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

// if in completeable state then return empty
// getCompletableCommand
//   - found then add to cache
// getCompleteFlag
//   - found then add to cache
// return empty suggestions if found either one

// First check in completeable state
// check if in flag that accepts arg state
// if flag accept arg
//   - check if parent is server completable flag
//   - check if that parent has this as child
//   - if yes then go ahead and look up the chached suggestions of the parent
// else
//   - find completable parent
//   - if found then get cached suggestions

// Afterwards look into the can accept more arg check thingy to see what they hell it is all about
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

	// if updated then it is a perent and just needs to update the cache for the completion of its children
	// suggestions only shown on its children
	if c.updatedCacheForCompletableCommandOrFlag(cmd) {
		return []prompt.Suggest{}
	}

	// "" if not in flag with arg state
	flagName := c.getFlagWithArg(d, cmd)
	if len(flagName) > 0 {
		suggestions := c.getSuggestionsForFlag(cmd, flagName)
		return filterSuggestions(d, suggestions)
	}

	//TODO: maybe do the num of arg check here?
	suggestions := c.getSuggestionsForCommand(d, cmd)
	return filterSuggestions(d, suggestions)
}

func (c *ServerSideCompleterImpl) getSuggestionsForCommand(d prompt.Document, cmd *cobra.Command) []prompt.Suggest {
	var suggestions []prompt.Suggest
	var cc ServerCompletableCommand
	if cc = c.getCompletableParent(cmd); cc == nil {
		return suggestions
	}
	suggestions, ok := c.getCachedSuggestions(d, cc)
	if !ok {
		// Shouldn't happen, but just in case.
		// If this does happen then cache should be in the process of updating.
		if !pcmd.CanCompleteCommand(cmd) {
			return suggestions
		}
		suggestions = cc.ServerComplete()
	}
	// if no suggestiosn just return empty list
	return suggestions
}

func (c *ServerSideCompleterImpl) updatedCacheForCompletableCommandOrFlag(cmd *cobra.Command) bool {
	var found bool
	v, ok := c.commandsByPath.Load(c.commandKey(cmd))
	if !ok {
		return false
	}
	var completableCommand ServerCompletableCommand
	var completableFlag ServerCompletableFlag
	if completableCommand, ok = v.(ServerCompletableCommand); ok {
		found = true
	}
	if completableFlag, ok = v.(ServerCompletableFlag); ok {
		found = true
	}
	go c.updateCachedSuggestions(completableCommand, completableFlag)
	return found
}

func (c *ServerSideCompleterImpl) getFlagWithArg(d prompt.Document, cmd *cobra.Command) string {
	promptArgs, _ := shellparser.Fields(d.TextBeforeCursor(), func(string) string { return "" })
	lastArg := promptArgs[len(promptArgs)-1]
	if !utils.IsFlagArg(lastArg) {
		return ""
	}

	flagList := c.getFlagList(cmd, promptArgs)
	return c.getFlagWithArgFromFlagList(flagList)
}

func (c *ServerSideCompleterImpl) getFlagList(cmd *cobra.Command, promptArgs []string) []*pflag.Flag {
	flagList := make([]*pflag.Flag, len(promptArgs))
	checkFlag := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		longName := "--" + flag.Name
		shortName := "-" + flag.Shorthand
		for i, arg := range promptArgs {
			if longName == arg || shortName == arg {
				flagList[i] = flag
			}
		}
	}

	cmd.LocalFlags().VisitAll(checkFlag)
	cmd.InheritedFlags().VisitAll(checkFlag)
	return flagList
}

func (c *ServerSideCompleterImpl) getFlagWithArgFromFlagList(flagList []*pflag.Flag) string {
	lastIndex := len(flagList) - 1
	if !isFlagWithArg(flagList[lastIndex]) {
		return ""
	}
	i := 0
	for ; i <= lastIndex; i += 1 {
		curFlag := flagList[i]
		if i == lastIndex && isFlagWithArg(curFlag) {
			return curFlag.Name
		}
		if isFlagWithArg(flagList[i]) {
			i += 1
		}
	}
	return ""
}

// is flag with arg
func isFlagWithArg(flag *pflag.Flag) bool {
	if flag == nil {
		return false
	}
	flagType := flag.Value.Type()
	return flagType != "bool" && flagType != "count"
}

// Check for flag suggestions, return empty list if not found or not in the state for flag suggestions
func (c *ServerSideCompleterImpl) getSuggestionsForFlag(cmd *cobra.Command, flagName string) []prompt.Suggest {

	// check static flag
	v, ok := c.staticFlagSuggestionsCommandMap.Load(flagName)
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
		// in case cahcing didn't finish
		return parent.ServerFlagComplete()[flagName]()
	}
	return v.([]prompt.Suggest)
}

// Check for flag suggestions, return empty list if not found or not in the state for flag suggestions
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

func (c *ServerSideCompleterImpl) updateCachedSuggestions(cc ServerCompletableCommand, cf ServerCompletableFlag) {
	canComplete := pcmd.CanCompleteCommand(cc.Cmd())
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

func (c *ServerSideCompleterImpl) updateCachedFlagSuggestions(cmd *cobra.Command, fc ServerCompletableFlag, wg *sync.WaitGroup) {
	for flagName, completeFunc := range fc.ServerFlagComplete() {
		c.cachedSuggestionsByPath.Store(c.flagKey(cmd, flagName), completeFunc())
	}
}

func (c *ServerSideCompleterImpl) getCachedSuggestions(d prompt.Document, cc ServerCompletableCommand) ([]prompt.Suggest, bool) {
	var suggestions []prompt.Suggest
	cmd := cc.Cmd()
	if fc, ok := cc.(ServerCompletableFlag); ok {
		// TODO: check that the command is the child
		_ = fc.ServerCompletableFlagChildren()
		findCompletableFlag := func(flag *pflag.Flag) {
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
				key := c.commandKey(cc.Cmd())
				v, ok := c.cachedSuggestionsByPath.Load(key)
				if ok {
					suggestions = v.([]prompt.Suggest)
				}
			}
		}
		cmd.LocalFlags().VisitAll(findCompletableFlag)
		cmd.InheritedFlags().VisitAll(findCompletableFlag)
		if len(suggestions) > 0 {
			return suggestions, true
		}
	}
	key := c.commandKey(cc.Cmd())
	v, ok := c.cachedSuggestionsByPath.Load(key)
	if !ok {
		return nil, false
	}
	return v.([]prompt.Suggest), true
}

func (c *ServerSideCompleterImpl) getFlagCachedSuggestions(fc ServerCompletableFlag) []prompt.Suggest {
	var suggestions []prompt.Suggest
	addFlags := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		longName := "--" + flag.Name
		if flag.DefValue == "" || flag.DefValue == "0" {
			key := c.flagKey(fc.Cmd(), flag.Name)
			v, ok := c.cachedSuggestionsByPath.Load(key)
			if !ok {
				// in case cache didn't finish...
				f, ok := fc.ServerFlagComplete()[longName]
				if ok {
					suggestions = f()
				}
			}
			suggestions = v.([]prompt.Suggest)
		}
	}
	cmd := fc.Cmd()
	cmd.LocalFlags().VisitAll(addFlags)
	cmd.InheritedFlags().VisitAll(addFlags)
	return suggestions
}

func (c *ServerSideCompleterImpl) getCachedFlagSuggestions(cmd *cobra.Command, flagName string) ([]prompt.Suggest, bool) {
	key := c.flagKey(cmd, flagName)
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

// getCompletableFlag returns a matching ServerCompletableFlag, or nil if one is not found.
func (c *ServerSideCompleterImpl) getCompletableFlags(cmd *cobra.Command) map[string]ServerCompletableFlag {
	var completableFlagMap map[string]ServerCompletableFlag
	addFlags := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
		}
		if flag.Hidden {
			return
		}
		longName := "--" + flag.Name
		if flag.DefValue == "" || flag.DefValue == "0" {
			v, ok := c.commandsByPath.Load(c.flagKey(cmd, flag.Name))
			if ok {
				completableFlagMap[longName] = v.(ServerCompletableFlag)
			}
		}
	}
	cmd.LocalFlags().VisitAll(addFlags)
	cmd.InheritedFlags().VisitAll(addFlags)
	return completableFlagMap
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

// getCompletableFlagParent return the ServerCompletableFlag parent if the specified command is a completable child,
// and false otherwise.
//func (c *ServerSideCompleterImpl) getCompletableFlagParent(d prompt.Document, cmd *cobra.Command) ServerCompletableFlag {
//	parent := cmd.Parent()
//	if parent == nil {
//		return nil
//	}
//	var completableFlag ServerCompletableFlag
//	findCompletableFlag := func(flag *pflag.Flag) {
//		if flag.Changed {
//			_ = flag.Value.Set(flag.DefValue)
//		}
//		if flag.Hidden {
//			return
//		}
//		longName := "--" + flag.Name
//		shortName := "-" + flag.Shorthand
//		endsWithFlag := strings.HasSuffix(d.GetWordBeforeCursorWithSpace(), shortName+" ") ||
//			strings.HasSuffix(d.GetWordBeforeCursorWithSpace(), longName+" ")
//		if endsWithFlag {
//			// should not suggest an argument if flag is not completed with a value but expects one
//			if flag.DefValue == "" || flag.DefValue == "0" {
//				if v, ok := fcs[longName]; ok {
//					completableFlag = v
//				}
//			}
//		}
//	}
//
//	cmd.LocalFlags().VisitAll(findCompletableFlag)
//	cmd.InheritedFlags().VisitAll(findCompletableFlag)
//	return completableFlag
//}

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

func (c *ServerSideCompleterImpl) AddFlag(cmd ServerCompletableFlag, flagName string) {
	c.commandsByPath.Store(c.flagKey(cmd.Cmd(), flagName), cmd)
}

func (c *ServerSideCompleterImpl) commandKey(cmd *cobra.Command) string {
	// trim CLI name
	return strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()+" ")
}

func (c *ServerSideCompleterImpl) flagKey(cmd *cobra.Command, flagName string) string {
	commandName := strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()+" ")
	return commandName + " --" + flagName
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
	//	const maxReasonableArgs = 20
	//	canAcceptMoreArgs := false
	//	for i := len(args) + 1; i <= maxReasonableArgs; i++ {
	//		tmpArgs := make([]string, i)
	//		if err := matchedCmd.ValidateArgs(tmpArgs); err == nil {
	//			canAcceptMoreArgs = true
	//			break
	//		}
	//	}
	//	if !canAcceptMoreArgs {
	//		fmt.Println("CNT ACCEPT MORE ARGS")
	//		return false
	//	}

	_ = matchedCmd.ParseFlags(strings.Fields(d.CurrentLine()))

	return shouldSuggest
}
