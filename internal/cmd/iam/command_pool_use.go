package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *identityPoolCommand) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id|name>",
		Short:             "Choose an identity pool to be used in subsequent commands.",
		Long:              "Choose an identity pool to be used in subsequent commands which support passing an identity pool with the `--identity-pool` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *identityPoolCommand) use(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	poolId := args[0]
	if _, err = c.V2Client.GetIdentityPool(poolId, provider); err != nil {
		if poolId, provider, err = c.poolAndProviderNamesToIds(poolId, provider); err != nil {
			return err
		}
		if _, err = c.V2Client.GetIdentityPool(poolId, provider); err != nil {
			return errors.NewErrorWithSuggestions(err.Error(), "List available identity pools with `confluent iam pool list`.")
		}
	}

	if err := c.Context.SetCurrentIdentityPool(poolId); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(errors.UsingResourceMsg, resource.IdentityPool, args[0])
	return nil
}
