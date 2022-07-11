package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
	if resource.LookupType(serviceAccountId) != resource.ServiceAccount {
		return errors.New(errors.BadServiceAccountIDErrorMsg)
	}

	sa, httpResp, err := c.V2Client.GetIamServiceAccount(serviceAccountId)
	if err != nil {
		return errors.CatchServiceAccountNotFoundError(err, httpResp, serviceAccountId)
	}

	return output.DescribeObject(cmd, &serviceAccount{
		ResourceId:  *sa.Id,
		Name:		 *sa.DisplayName,
		Description: *sa.Description,
	}, describeFields, describeHumanRenames, describeStructuredRenames)
}
