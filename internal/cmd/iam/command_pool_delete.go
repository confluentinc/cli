package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *identityPoolCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete identity pools.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete identity pool "pool-12345":`,
				Code: "confluent iam pool delete pool-12345 --provider op-12345",
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *identityPoolCommand) delete(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	if err := c.confirmDeletion(cmd, provider, args); err != nil {
		return err
	}

	deleted, err := resource.Delete(args, func(id string) error {
		if err := c.V2Client.DeleteIdentityPool(id, provider); err != nil {
			return err
		}
		return nil
	}, resource.DefaultPostProcess)
	resource.PrintDeleteSuccessMsg(deleted, resource.IdentityPool)

	return err
}

func (c *identityPoolCommand) confirmDeletion(cmd *cobra.Command, provider string, args []string) error {
	var displayName string
	describeFunc := func(id string) error {
		pool, err := c.V2Client.GetIdentityPool(id, provider)
		if err == nil && id == args[0] {
			displayName = pool.GetDisplayName()
		}
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.IdentityPool, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.IdentityPool, args[0], displayName), displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.IdentityPool, args)); err != nil || !ok {
			return err
		}
	}

	return nil
}
