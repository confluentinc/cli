package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type offsetStatusCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newStatusCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of offset update",
	}

	c := &offsetStatusCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.AddCommand(c.newStatusDescribeCommand())

	return cmd
}
