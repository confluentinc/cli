package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type shareCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newShareCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage Kafka shares.",
	}

	c := &shareCommand{}

	// Only cloud support for now
	c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	cmd.AddCommand(c.newGroupCommand())

	return cmd
}
