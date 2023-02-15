package byok

import (
	"errors"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	errorMsgs "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List self-managed keys.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddByokProviderFlag(cmd)
	pcmd.AddByokStateFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {

	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}
	switch provider {
	case "aws":
		provider = "AWS"
	case "azure":
		provider = "Azure"
	}

	state, err := cmd.Flags().GetString("state")
	if err != nil {
		return err
	}
	switch state {
	case "in-use":
		state = "IN_USE"
	case "available":
		state = "AVAILABLE"
	}

	keys, err := c.V2Client.ListByokKeys(provider, state)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	// API returns a list sorted by creation date already
	list.Sort(false)
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
			Id:        key.GetId(),
			Key:       keyString,
			Provider:  key.GetProvider(),
			State:     key.GetState(),
			CreatedAt: key.Metadata.CreatedAt.String(),
			UpdatedAt: updatedAt,
			DeletedAt: deletedAt,
		})

	}

	list.Filter([]string{"Id", "Key", "Provider", "State", "CreatedAt", "UpdatedAt"})
	return list.Print()
}
