package auditlog

import (
	"context"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type configCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newConfigCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage the audit log configuration specification.",
		Long:        "Manage the audit log defaults and routing rules that determine which auditable events are logged, and where.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &configCommand{pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)}

	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newEditCommand())
	c.AddCommand(c.newMigrateCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

func (c *configCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
}
