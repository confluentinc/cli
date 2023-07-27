package cmd

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func CatchErrors(fs ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		f := chain(fs...)
		if err := errors.HandleCommon(f(cmd, args)); err != nil {
			// Only show usage for Cobra-related errors (missing args, incorrect flags, etc.)
			cmd.SilenceUsage = true
			return err
		}
		return nil
	}
}

func chain(fs ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range fs {
			if err := f(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}

// NewValidArgsFunction is a wrapper around `cobra.ValidArgsFunction()` that ignores the `toComplete`
// argument and `ShellCompDirective` return value, which are almost always ignored and "NoFileComp", respectively.
func NewValidArgsFunction(f func(*cobra.Command, []string) []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		return f(cmd, args), cobra.ShellCompDirectiveNoFileComp
	}
}

// RegisterFlagCompletionFunc is a wrapper around `cobra.RegisterFlagCompletionFunc()` that ignores the `toComplete`
// argument and `ShellCompDirective` return value, which are almost always ignored and "NoFileComp", respectively.
func RegisterFlagCompletionFunc(cmd *cobra.Command, flag string, f func(*cobra.Command, []string) []string) {
	_ = cmd.RegisterFlagCompletionFunc(flag, func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		return f(cmd, args), cobra.ShellCompDirectiveNoFileComp
	})
}
