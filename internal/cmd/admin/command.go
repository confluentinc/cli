package admin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	isTest bool
}

func New(prerunner pcmd.PreRunner, isTest bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "admin",
		Short:       "Perform administrative tasks for the current organization.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		isTest:                  isTest,
	}

	c.AddCommand(c.newPaymentCommand())
	c.AddCommand(c.newPromoCommand())

	return c.Command
}
