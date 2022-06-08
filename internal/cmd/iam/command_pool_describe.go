package iam

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	poolHumanLabelMap = map[string]string{
		"Id":           "ID",
		"DisplayName":  "Display Name",
		"Description":  "Description",
		"SubjectClaim": "Subject Claim",
	}
	poolStructuredLabelMap = map[string]string{
		"Id":           "id",
		"DisplayName":  "display_name",
		"Description":  "description",
		"SubjectClaim": "subject_claim",
	}
)

func (c identityPoolCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe an identity pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	_ = cmd.MarkFlagRequired("provider")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c identityPoolCommand) describe(cmd *cobra.Command, args []string) error {
	if resource.LookupType(args[0]) != resource.IdentityPool {
		return fmt.Errorf(errors.BadResourceIDErrorMsg, resource.IdentityPoolPrefix)
	}

	providerID, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityPoolProfile, _, err := c.V2Client.GetIdentityPool(providerID, args[0])
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, &identityPool{
		Id:           *identityPoolProfile.Id,
		DisplayName:  *identityPoolProfile.DisplayName,
		Description:  *identityPoolProfile.Description,
		SubjectClaim: *identityPoolProfile.SubjectClaim,
	}, poolListFields, poolHumanLabelMap, poolStructuredLabelMap)
}
