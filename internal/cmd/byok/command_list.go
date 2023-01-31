package byok

import (
	"errors"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	errorMsgs "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var listFields = []string{"Id", "Key", "Provider", "State", "CreatedAt", "UpdatedAt", "DeletedAt"}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List self-managed keys.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {

	keys, err := c.V2Client.ListByokKeys()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, key := range keys {
		var keyString string
		switch {
		case key.Key.ByokV1AwsKey != nil:
			keyString = key.Key.ByokV1AwsKey.KeyArn
		case key.Key.ByokV1AzureKey != nil:
			keyString = key.Key.ByokV1AzureKey.KeyId
		default:
			return errors.New(errorMsgs.ByokUnknownKeyTypeErrorMsg)
		}

		updatedAt := ""
		if !key.Metadata.GetUpdatedAt().IsZero() {
			updatedAt = key.Metadata.GetUpdatedAt().String()
		}

		deletedAt := ""
		if !key.Metadata.GetDeletedAt().IsZero() {
			deletedAt = key.Metadata.GetDeletedAt().String()
		}

		list.Add(&describeStruct{
			Id:        *key.Id,
			Key:       keyString,
			Provider:  *key.Provider,
			State:     *key.State,
			CreatedAt: key.Metadata.CreatedAt.String(),
			UpdatedAt: updatedAt,
			DeletedAt: deletedAt,
		})

	}

	list.Filter(listFields)
	return list.Print()

}
