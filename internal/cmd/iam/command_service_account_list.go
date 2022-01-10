package iam

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	serviceAccountListFields           = []string{"ResourceId", "ServiceName", "ServiceDescription"}
	serviceAccountListHumanLabels      = []string{"ID", "Name", "Description"}
	serviceAccountListStructuredLabels = []string{"id", "name", "description"}
)

func (c *serviceAccountCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) list(cmd *cobra.Command, _ []string) error {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, serviceAccountListFields, serviceAccountListHumanLabels, serviceAccountListStructuredLabels)
	if err != nil {
		return err
	}
	for _, u := range users {
		outputWriter.AddElement(u)
	}
	return outputWriter.Out()
}
