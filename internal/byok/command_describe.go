package byok

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a self-managed key.",
		Long:              "Describe a self-managed key in Confluent Cloud.",
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
		return errors.CatchByokKeyNotFoundError(err, httpResp)
	}
	return c.outputByokKeyDescription(cmd, key)
}

func (c *command) outputByokKeyDescription(cmd *cobra.Command, key byokv1.ByokV1Key) error {
	var keyString string
	var roles []string

	switch {
	case key.Key.ByokV1AwsKey != nil:
		keyString = key.Key.ByokV1AwsKey.KeyArn
		roles = key.Key.ByokV1AwsKey.GetRoles()
	case key.Key.ByokV1AzureKey != nil:
		keyString = key.Key.ByokV1AzureKey.KeyId
		roles = append(roles, key.Key.ByokV1AzureKey.GetApplicationId())
	case key.Key.ByokV1GcpKey != nil:
		keyString = key.Key.ByokV1GcpKey.KeyId
		roles = append(roles, key.Key.ByokV1GcpKey.GetSecurityGroup())
	default:
		return fmt.Errorf(byokUnknownKeyTypeErrorMsg)
	}

	table := output.NewTable(cmd)
	if output.GetFormat(cmd) == output.Human {
		table.Add(&humanOut{
			Id:        key.GetId(),
			Key:       keyString,
			Roles:     strings.Join(roles, ", "),
			Provider:  key.GetProvider(),
			State:     key.GetState(),
			CreatedAt: key.Metadata.CreatedAt.String(),
		})
	} else {
		table.Add(&serializedOut{
			Id:        key.GetId(),
			Key:       keyString,
			Roles:     roles,
			Provider:  key.GetProvider(),
			State:     key.GetState(),
			CreatedAt: key.Metadata.CreatedAt.String(),
		})
	}
	table.Print()

	if output.GetFormat(cmd) == output.Human {
		postCreationStepInstructions, err := getPolicyCommand(key)
		if err != nil {
			return err
		}

		output.ErrPrintln(c.Config.EnableColor, "")
		output.ErrPrintln(c.Config.EnableColor, getPostCreateStepInstruction(key))
		output.ErrPrintln(c.Config.EnableColor, "")
		output.ErrPrintln(c.Config.EnableColor, postCreationStepInstructions)
	}

	return nil
}
