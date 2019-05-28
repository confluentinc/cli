package lint_cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RuleFilter returns true if linter should be run on command
type RuleFilter func(*cobra.Command) bool

// FlagRuleFilter returns true if linter should be run on command flag
type FlagRuleFilter func(*pflag.Flag, *cobra.Command) bool

func Filter(rule Rule, filters ...RuleFilter) Rule {
	return func(cmd *cobra.Command) error {
		for _, f := range filters {
			if !f(cmd) {
				return nil
			}
		}
		return rule(cmd)
	}
}

func FlagFilter(rule FlagRule, filters ...FlagRuleFilter) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		for _, f := range filters {
			if !f(flag, cmd) {
				return nil
			}
		}
		return rule(flag, cmd)
	}
}

func OnlyLeafCommands(cmd *cobra.Command) bool {
	return !cmd.HasAvailableSubCommands()
}

func ExcludeCommand(excluded ...string) RuleFilter {
	blacklist := map[string]struct{}{}
	for _, e := range excluded {
		blacklist[e] = struct{}{}
	}
	return func(cmd *cobra.Command) bool {
		f := FullCommand(cmd)
		exclude := strings.Join(strings.Split(f, " ")[1:], " ")
		if _, found := blacklist[exclude]; found {
			return false
		}
		return true
	}
}

func ExcludeUse(excluded ...string) RuleFilter {
	blacklist := map[string]struct{}{}
	for _, e := range excluded {
		blacklist[e] = struct{}{}
	}
	return func(cmd *cobra.Command) bool {
		if _, found := blacklist[cmd.Use]; found {
			return false
		}
		return true
	}
}

func ExcludeParentUse(excluded ...string) RuleFilter {
	blacklist := map[string]struct{}{}
	for _, e := range excluded {
		blacklist[e] = struct{}{}
	}
	return func(cmd *cobra.Command) bool {
		if _, found := blacklist[cmd.Parent().Use]; found {
			return false
		}
		return true
	}
}

// ExcludeCommandContains specifies a blacklist of commands to which this rule does not apply
func ExcludeCommandContains(excluded ...string) RuleFilter {
	return func(cmd *cobra.Command) bool {
		exclude := true
		for _, ex := range excluded {
			if strings.Contains(FullCommand(cmd), ex) {
				exclude = false
				break
			}
		}
		return exclude
	}
}

// IncludeCommandContains specifies a whitelist of commands to which this rule applies
func IncludeCommandContains(included ...string) RuleFilter {
	return func(cmd *cobra.Command) bool {
		include := false
		for _, in := range included {
			if strings.Contains(FullCommand(cmd), in) {
				include = true
				break
			}
		}
		return include
	}
}

// Exclude flags by name
func ExcludeFlag(excluded ...string) FlagRuleFilter {
	blacklist := map[string]struct{}{}
	for _, e := range excluded {
		blacklist[e] = struct{}{}
	}
	return func(flag *pflag.Flag, cmd *cobra.Command) bool {
		if _, found := blacklist[flag.Name]; found {
			return false
		}
		return true
	}
}
