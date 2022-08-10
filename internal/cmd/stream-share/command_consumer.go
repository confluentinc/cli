package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type consumerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newConsumerCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage consumer actions.",
	}

	c := &consumerCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(newConsumerShareCommand(prerunner))
	c.AddCommand(c.newRedeemCommand())

	return c.Command
}
