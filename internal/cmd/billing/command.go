package billing

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "billing",
		Short:       "Manage Confluent Cloud Billing.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}
	c := &commandCost{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.Command.AddCommand(c.newCostCommand())

	return c.Command
}
