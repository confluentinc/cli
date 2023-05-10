package billing

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func newCostCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Manage Confluent Cloud Billing Costs.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.AddCommand(c.newListCostCommand())

	return cmd
}
