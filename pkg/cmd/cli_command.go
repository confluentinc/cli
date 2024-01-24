package cmd

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/version"
)

type CLICommand struct {
	*cobra.Command
	Config  *config.Config
	Version *version.Version
}

func NewCLICommand(cmd *cobra.Command) *CLICommand {
	return &CLICommand{
		Command: cmd,
		Config:  &config.Config{},
	}
}

func NewAnonymousCLICommand(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	c := NewCLICommand(cmd)
	cmd.PersistentPreRunE = chain(prerunner.Anonymous(c, false), prerunner.ParseFlagsIntoContext(c))
	return c
}
