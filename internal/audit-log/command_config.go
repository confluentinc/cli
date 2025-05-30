package auditlog

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type configCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newConfigCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage the audit log configuration specification.",
		Long:        "Manage the audit log defaults and routing rules that determine which auditable events are logged, and where.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &configCommand{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newEditCommand())
	cmd.AddCommand(c.newMigrateCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *configCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
}
