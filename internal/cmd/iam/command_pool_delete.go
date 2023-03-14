package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *identityPoolCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete identity pools.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
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

	_ = cmd.MarkFlagRequired("provider")

	return cmd
}

func (c *identityPoolCommand) delete(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	displayName, err := c.checkExistence(cmd, provider, args)
	if err != nil {
		return err
	}

	if _, err := form.ConfirmDeletionType(cmd, resource.IdentityPool, displayName, args); err != nil {
		return err
	}

	var errs error
	for _, poolId := range args {
		if err := c.V2Client.DeleteIdentityPool(poolId, provider); err != nil {
			errs = errors.Join(errs, err)
		} else {
			output.ErrPrintf(errors.DeletedResourceMsg, resource.IdentityPool, poolId)
		}
	}

	return errs
}

func (c *identityPoolCommand) checkExistence(cmd *cobra.Command, provider string, args []string) (string, error) {
	// Single
	if len(args) == 1 {
		if pool, err := c.V2Client.GetIdentityPool(args[0], provider); err != nil {
			return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.IdentityPool, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.IdentityPool))
		} else {
			return pool.GetDisplayName(), nil
		}
	}

	// Multiple
	identityPools, err := c.V2Client.ListIdentityPools(provider)
	if err != nil {
		return "", err
	}

	set := types.NewSet()
	for _, pool := range identityPools {
		set.Add(pool.GetId())
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return "", err
	} else if force && len(invalidArgs) > 0 {
		args = validArgs
		return "", nil
	}

	invalidArgsStr := utils.ArrayToCommaDelimitedStringWithAnd(invalidArgs)
	if len(invalidArgs) == 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.IdentityPool, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.IdentityPool))
	} else if len(invalidArgs) > 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resource.IdentityPool), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.IdentityPool))
	}

	return "", nil
}
