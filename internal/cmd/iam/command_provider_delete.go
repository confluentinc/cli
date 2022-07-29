package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *identityProviderCommand) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an identity provider.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete identity provider "op-12345":`,
				Code: "confluent iam provider delete op-12345",
			},
		),
	}
}

func (c *identityProviderCommand) delete(cmd *cobra.Command, args []string) error {
	if httpResp, err := c.V2Client.DeleteIdentityProvider(args[0]); err != nil {
		return errors.CatchV2ErrorMessageWithResponse(err, httpResp)
	}

	utils.ErrPrintf(cmd, errors.DeletedIdentityProviderMsg, args[0])
	return nil
}
