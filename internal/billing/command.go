package billing

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "billing",
		Short:       "Manage Confluent Cloud billing.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginAllowFreeTrialEnded},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCostCommand())
	cmd.AddCommand(c.newPaymentCommand())
	cmd.AddCommand(c.newPriceCommand())
	cmd.AddCommand(c.newPromoCommand())

	return cmd
}
