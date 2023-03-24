package local

import (
	"runtime"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Manage a local Confluent Platform development environment.",
		Long:  `Use the "confluent local" commands to try out Confluent Platform by running a single-node instance locally on your machine. These commands require Docker to run.`,
		Args:  cobra.NoArgs,
	}

	if runtime.GOOS == "windows" {
		cmd.Hidden = true
	}

	c := &command{pcmd.NewAnonymousCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newKafkaCommand())

	return cmd
}
