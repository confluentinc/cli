package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c serviceAccountCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a service account.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c serviceAccountCommand) describe(cmd *cobra.Command, args []string) error {
	serviceAccountId := args[0]

	serviceAccount, httpResp, err := c.V2Client.GetIamServiceAccount(serviceAccountId)
	if err != nil {
		return errors.CatchServiceAccountNotFoundError(err, httpResp, serviceAccountId)
	}

	table := output.NewTable(cmd)
	table.Add(&serviceAccountOut{
		ResourceId:  serviceAccount.GetId(),
		Name:        serviceAccount.GetDisplayName(),
		Description: serviceAccount.GetDescription(),
	})
	return table.Print()
}
