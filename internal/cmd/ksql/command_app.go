package ksql

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
)

func newAppCommand(prerunner pcmd.PreRunner) *ksqlCommand {
	cmd := &cobra.Command{
		Use:         "app",
		Short:       "DEPRECATED: Manage ksqlDB apps.",
		Long:        "DEPRECATED: Manage ksqlDB apps. " + errors.KSQLAppDeprecateWarning,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ksqlCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),}

	c.AddCommand(c.newConfigureAclsCommand(true))
	c.AddCommand(c.newCreateCommand(true))
	c.AddCommand(c.newDeleteCommand(true))
	c.AddCommand(c.newDescribeCommand(true))
	c.AddCommand(c.newListCommand(true))

	return c
}