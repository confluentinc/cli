package admin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner, isTest bool) *cobra.Command {
	c := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:         "admin",
			Short:       "Perform administrative tasks for the current organization.",
			Args:        cobra.NoArgs,
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		},
		prerunner,
	)

	c.AddCommand(NewPaymentCommand(prerunner, isTest))
	c.AddCommand(NewPromoCommand(prerunner))

	return c.Command
}
