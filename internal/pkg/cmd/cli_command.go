package cmd

import (
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

type CLICommand struct {
	*cobra.Command
	Config    *dynamicconfig.DynamicConfig
	Version   *version.Version
	prerunner PreRunner
}

func NewCLICommand(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	return &CLICommand{
		Config:    &dynamicconfig.DynamicConfig{},
		Command:   cmd,
		prerunner: prerunner,
	}
}

func NewAnonymousCLICommand(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	c := NewCLICommand(cmd, prerunner)
	cmd.PersistentPreRunE = Chain(prerunner.Anonymous(c, false), prerunner.AnonymousParseFlagsIntoContext(c))
	c.Command = cmd
	return c
}
