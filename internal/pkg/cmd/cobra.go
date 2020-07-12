package cmd

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func NewCLIRunE(runEFunc func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return errors.HandleCommon(runEFunc(cmd, args), cmd)
	}
}

func NewCLIPreRunnerE(prerunnerE func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return errors.HandleCommon(prerunnerE(cmd, args), cmd)
	}
}
