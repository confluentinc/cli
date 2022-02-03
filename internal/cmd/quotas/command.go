package quotas

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "applied-quotas",
		Aliases:     []string{"quota"},
		Short:       "Look up ccloud service quotas limits",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newListCmd())

	return c.Command
}
