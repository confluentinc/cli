package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	poolHumanLabelMap = map[string]string{
		"Id":            "ID",
		"DisplayName":   "Display Name",
		"IdentityClaim": "Identity Claim",
	}
	poolStructuredLabelMap = map[string]string{
		"Id":            "id",
		"DisplayName":   "display_name",
		"Description":   "description",
		"IdentityClaim": "identity_claim",
		"Filter":        "filter",
	}
)

func (c identityPoolCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an identity pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("provider")
	return cmd
}

func (c identityPoolCommand) describe(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityPoolProfile, httpResp, err := c.V2Client.GetIdentityPool(args[0], provider)
	if err != nil {
		return errors.CatchV2ErrorMessageWithResponse(err, httpResp)
	}

	describeIdentityPool := &identityPool{
		Id:            *identityPoolProfile.Id,
		DisplayName:   *identityPoolProfile.DisplayName,
		IdentityClaim: *identityPoolProfile.IdentityClaim,
		Filter:        *identityPoolProfile.Filter,
	}
	if identityPoolProfile.Description != nil {
		describeIdentityPool.Description = *identityPoolProfile.Description
	}

	return output.DescribeObject(cmd, describeIdentityPool, identityPoolListFields, poolHumanLabelMap, poolStructuredLabelMap)
}
