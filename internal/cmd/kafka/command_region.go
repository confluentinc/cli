package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newRegionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "region",
		Short:       "Manage Confluent Cloud regions.",
		Long:        "Use this command to manage Confluent Cloud regions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &regionCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newListCommand())

	return cmd
}
