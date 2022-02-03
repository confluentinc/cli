package linter

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CommandRuleFilter func(*cobra.Command) bool
	FlagRuleFilter    func(*pflag.Flag, *cobra.Command) bool
)

func Filter(rule CommandRule, filters ...CommandRuleFilter) CommandRule {
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

func ExcludeCommand(commands ...string) CommandRuleFilter {
	blacklist := map[string]struct{}{}
	for _, command := range commands {
		blacklist[command] = struct{}{}
	}

	return func(cmd *cobra.Command) bool {
		command := strings.TrimPrefix(cmd.CommandPath(), "confluent ")
		_, ok := blacklist[command]
		return !ok
	}
}

// ExcludeCommandContains specifies a blacklist of commands to which this rule does not apply
func ExcludeCommandContains(excluded ...string) CommandRuleFilter {
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

// ExcludeFlag excludes flags by name
func ExcludeFlag(excluded ...string) FlagRuleFilter {
	blacklist := map[string]struct{}{}
	for _, e := range excluded {
		blacklist[e] = struct{}{}
	}
	return func(flag *pflag.Flag, cmd *cobra.Command) bool {
		_, found := blacklist[flag.Name]
		return !found
	}
}
