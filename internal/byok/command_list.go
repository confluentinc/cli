package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List self-managed keys.",
		Long:  "List self-managed keys registered in Confluent Cloud.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddByokStateFlag(cmd)
	cmd.Flags().String("region", "", "Filter by region.")
	cmd.Flags().String("phase", "", "Filter by validation phase.")
	cmd.Flags().String("display-name", "", "Filter by display name.")
	cmd.Flags().String("key", "", "Filter by key identifier.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	switch cloud {
	case "aws":
		cloud = "AWS"
	case "azure":
		cloud = "Azure"
	case "gcp":
		cloud = "GCP"
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

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetString("phase")
	if err != nil {
		return err
	}

	displayName, err := cmd.Flags().GetString("display-name")
	if err != nil {
		return err
	}

	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}

	keys, err := c.V2Client.ListByokKeys(cloud, state, region, phase, displayName, key)
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
		case key.Key.ByokV1GcpKey != nil:
			keyString = key.Key.ByokV1GcpKey.KeyId
		default:
			return errors.New(byokUnknownKeyTypeErrorMsg)
		}

		list.Add(&out{
			Id:              key.GetId(),
			DisplayName:     key.GetDisplayName(),
			Key:             keyString,
			Cloud:           key.GetProvider(),
			State:           key.GetState(),
			CreatedAt:       key.Metadata.CreatedAt.String(),
			ValidationPhase: key.Validation.GetPhase(),
		})
	}

	// The API returns a list sorted by creation date already
	list.Sort(false)
	list.Filter([]string{"Id", "DisplayName", "Key", "Cloud", "State", "ValidationPhase", "CreatedAt"})

	return list.Print()
}
