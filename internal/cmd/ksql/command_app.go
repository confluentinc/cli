package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newAppCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "app",
		Short:       "DEPRECATED: Manage ksqlDB apps.",
		Long:        "DEPRECATED: Manage ksqlDB apps. " + errors.KSQLAppDeprecateWarning,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ksqlCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newConfigureAclsCommand(true))
	cmd.AddCommand(c.newCreateCommand(true))
	cmd.AddCommand(c.newDeleteCommand(true))
	cmd.AddCommand(c.newDescribeCommand(true))
	cmd.AddCommand(c.newListCommand(true))

	return cmd
}
