package ksql

import (
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
)

func newAppCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *ksqlCommand {
	cmd := &cobra.Command{
		Use:         "app",
		Short:       "Manage ksqlDB apps. " + errors.KSQLAppDeprecateWarning,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ksqlCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		analyticsClient:               analyticsClient,
	}

	c.AddCommand(c.newConfigureAclsCommand(true))
	c.AddCommand(c.newCreateCommand(true))
	c.AddCommand(c.newDeleteCommand(true))
	c.AddCommand(c.newDescribeCommand(true))
	c.AddCommand(c.newListCommand(true))

	return c
}