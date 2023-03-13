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
		Use:               "delete <id-1> [id-2] ... [id-N]",
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
			return "", err
		} else {
			return pool.GetDisplayName(), nil
		}
	}

	// Multiple
	identityPools, err := c.V2Client.ListIdentityPools(provider)
	if err != nil {
		return "", err
	}

	poolSet := types.NewSet()
	for _, pool := range identityPools {
		poolSet.Add(pool.GetId())
	}

	invalidPools := poolSet.Difference(args)
	if len(invalidPools) > 0 {
		return "", errors.New(fmt.Sprintf(errors.AccessForbiddenErrorMsg, resource.IdentityPool, utils.ArrayToCommaDelimitedStringWithAnd(invalidPools)))
	}

	return "", nil
}
