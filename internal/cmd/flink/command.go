package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flink",
		Short: "Manage Apache Flink.",
	}

	c := &command{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newStatementCommand())

	return cmd
}
