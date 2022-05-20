package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type quotaCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newQuotaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "quota",
		Short:       "Manage Kafka client quotas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &quotaCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newCreateCommand())
	//c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newListCommand())
	//c.AddCommand(c.newUpdateCommand())

	return c.Command
}
