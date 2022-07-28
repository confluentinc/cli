package streamshare

import (
	"github.com/spf13/cobra"
	
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type providerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newProviderCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "provider",
		Short:       "Manage provider actions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &providerCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(newProviderShareCommand(prerunner))

	return c.Command
}
