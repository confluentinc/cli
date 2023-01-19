package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a self-managed key in Confluent Cloud.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	key, httpResp, err := c.V2Client.GetByokKey(args[0])
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, httpResp)
	}

	var keyString string
	var roles []string
	switch {
	case key.Key.ByokV1AwsKey != nil:
		keyString = key.Key.ByokV1AwsKey.KeyArn
		roles = key.Key.ByokV1AwsKey.GetRoles()
	case key.Key.ByokV1AzureKey != nil:
		keyString = key.Key.ByokV1AzureKey.KeyId
		roles = append(roles, key.Key.ByokV1AzureKey.GetApplicationId())
	default:
		return errors.New("unknown key type")
	}
	var updatedAt string
	if key.Metadata.UpdatedAt == nil || key.Metadata.UpdatedAt.IsZero() {
		updatedAt = ""
	} else {
		updatedAt = key.Metadata.UpdatedAt.String()
	}
	var deletedAt string
	if key.Metadata.DeletedAt == nil || key.Metadata.UpdatedAt.IsZero() {
		deletedAt = ""
	} else {
		deletedAt = key.Metadata.DeletedAt.String()
	}

	describeByokKey := &byokKey{
		Id:        *key.Id,
		Key:       keyString,
		Roles:     roles,
		Provider:  *key.Provider,
		State:     *key.State,
		CreatedAt: key.Metadata.CreatedAt.String(),
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}

	return output.DescribeObject(cmd, describeByokKey, fields, humanRenames, structuredRenames)
}
