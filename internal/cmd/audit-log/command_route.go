package auditlog

import (
	"context"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type routeCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newRouteCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "route",
		Short:       "Return the audit log route rules.",
		Long:        "Return the routing rules that determine which auditable events are logged, and where.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &routeCommand{pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newLookupCommand())

	return cmd
}

func (c *routeCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
}
