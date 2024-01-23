package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *identityProviderCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more identity providers.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete identity provider "op-12345":`,
				Code: "confluent iam provider delete op-12345",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) delete(cmd *cobra.Command, args []string) error {
	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetIdentityProvider(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.IdentityProvider); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteIdentityProvider(id)
	}

	_, err := deletion.Delete(args, deleteFunc, resource.IdentityProvider)
	return err
}
