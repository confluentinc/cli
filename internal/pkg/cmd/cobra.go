package cmd

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

// NewCLIRunE - Wrapper function around RunE for formatting more helpful error messages when creating a cobra.Command
// see https://github.com/confluentinc/cli/blob/master/errors.md
func NewCLIRunE(runEFunc func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return errors.HandleCommon(runEFunc(cmd, args), cmd)
	}
}

// NewCLIPreRunnerE - Wrapper function around PreRunnerE for formatting more helpful error messages when creating a cobra.Command
func NewCLIPreRunnerE(prerunnerE ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, prerunner := range prerunnerE {
			err := prerunner(cmd, args)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
		}
		return nil
	}
}
