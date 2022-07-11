package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c serviceAccountCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c serviceAccountCommand) describe(cmd *cobra.Command, args []string) error {
	serviceAccountId := args[0]

	sa, httpResp, err := c.V2Client.GetIamServiceAccount(serviceAccountId)
	if err != nil {
		return errors.CatchServiceAccountNotFoundError(err, httpResp, serviceAccountId)
	}

	saOut := &serviceAccount{
		ResourceId:  *sa.Id,
		Name:		 *sa.DisplayName,
		Description: *sa.Description,
	}
	return output.DescribeObject(cmd, saOut, describeFields, describeHumanRenames, describeStructuredRenames)
}
