package org

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage Organization.",
	}

	cmd.AddCommand(
		newScimTokenCommand(cfg, prerunner),
		// cli-tfgen:cli-subcommands
	)

	return cmd
}
