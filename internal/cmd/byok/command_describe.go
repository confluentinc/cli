package byok

import (
	"fmt"

	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
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

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	key, httpResp, err := c.V2Client.GetByokKey(args[0])
	if err != nil {
		return errors.CatchByokKeyNotFoundError(err, httpResp)
	}

	return c.outputByokKeyDescription(cmd, &key)
}

func (c *command) outputByokKeyDescription(cmd *cobra.Command, key *byokv1.ByokV1Key) error {
	postCreationStepInstructions, err := getPolicyCommand(key)
	if err != nil {
		return err
	}

	out, err := c.convertByokKeyToDescribeStruct(key)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter([]string{"Id", "Key", "Roles", "Provider", "State", "CreatedAt"})
	table.Print()

	if output.GetFormat(cmd) == output.Human {
		utils.ErrPrintln(cmd, fmt.Sprintf("\n%s\n", getPostCreateStepInstruction(key)))
		utils.Println(cmd, postCreationStepInstructions)
	}

	return nil
}

func (c *command) convertByokKeyToDescribeStruct(key *byokv1.ByokV1Key) (*describeStruct, error) {
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
		return nil, errors.New(errorMsg.ByokUnknownKeyTypeErrorMsg)
	}

	updatedAt := ""
	if !key.Metadata.GetUpdatedAt().IsZero() {
		updatedAt = key.Metadata.GetUpdatedAt().String()
	}

	deletedAt := ""
	if !key.Metadata.GetDeletedAt().IsZero() {
		deletedAt = key.Metadata.GetDeletedAt().String()
	}

	describeKey := &describeStruct{
		Id:        key.GetId(),
		Key:       keyString,
		Roles:     roles,
		Provider:  key.GetProvider(),
		State:     key.GetState(),
		CreatedAt: key.Metadata.CreatedAt.String(),
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}

	return describeKey, nil
}
