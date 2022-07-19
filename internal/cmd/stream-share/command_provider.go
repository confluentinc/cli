package stream_share

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
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

	s := &providerCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	s.AddCommand(newProviderShareCommand(prerunner))

	return s.Command
}
