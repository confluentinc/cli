package billing

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type commandCost struct {
	*pcmd.AuthenticatedCLICommand
}

func (c *commandCost) newCostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Manage Confluent Cloud Billing Costs.",
	}

	cl := &commandList{c.AuthenticatedCLICommand}
	cmd.AddCommand(cl.newCostListCommand())

	return cmd
}
