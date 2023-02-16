package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	errorMsg "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type describeStruct struct {
	Id        string   `human:"ID" serialized:"id"`
	Key       string   `human:"Key" serialized:"key"`
	Roles     []string `human:"Roles" serialized:"roles"`
	Provider  string   `human:"Provider" serialized:"provider"`
	State     string   `human:"State" serialized:"state"`
	CreatedAt string   `human:"Created At" serialized:"created_at"`
	UpdatedAt string   `human:"Updated At" serialized:"updated_at"`
	DeletedAt string   `human:"Deleted At" serialized:"deleted_at"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a self-managed key.",
		Long:              "Describe a self-managed key in Confluent Cloud.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	cmd.Flags().Bool("policy-command", false, "Print post-creation step to grant Confluent access to the key.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	key, httpResp, err := c.V2Client.GetByokKey(args[0])
	if err != nil {
		return errors.CatchByokKeyNotFoundError(err, httpResp)
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
		return errors.New(errorMsg.ByokUnknownKeyTypeErrorMsg)
	}

	updatedAt := ""
	if !key.Metadata.GetUpdatedAt().IsZero() {
		updatedAt = key.Metadata.GetUpdatedAt().String()
	}

	deletedAt := ""
	if !key.Metadata.GetDeletedAt().IsZero() {
		deletedAt = key.Metadata.GetDeletedAt().String()
	}

	table := output.NewTable(cmd)
	table.Add(&describeStruct{
		Id:        key.GetId(),
		Key:       keyString,
		Roles:     roles,
		Provider:  key.GetProvider(),
		State:     key.GetState(),
		CreatedAt: key.Metadata.CreatedAt.String(),
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	})
	table.Print()

	showPolicyCommand, err := cmd.Flags().GetBool("policy-command")
	if err != nil {
		return err
	}

	if showPolicyCommand {
		postCreationStepInstructions, err := getPostCreationStepInstructions(&key)
		if err != nil {
			return err
		}

		utils.Println(cmd, postCreationStepInstructions)
	}

	return nil
}
