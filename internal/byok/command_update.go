package byok

import (
	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a self-managed key.",
		Long:              "Update a self-managed key in Confluent Cloud.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the display name of self-managed key "cck-12345".`,
				Code: `confluent byok update cck-12345 --display-name "My production key"`,
			},
		),
	}

	cmd.Flags().String("display-name", "", "A human-readable name for the self-managed key.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("display-name"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	keyId := args[0]

	displayName, err := cmd.Flags().GetString("display-name")
	if err != nil {
		return err
	}

	// Use dedicated update struct following the pattern from API Keys
	updateReq := byokv1.ByokV1KeyUpdate{
		DisplayName: byokv1.PtrString(displayName),
	}

	key, httpResp, err := c.V2Client.UpdateByokKey(keyId, updateReq)
	if err != nil {
		return errors.CatchByokKeyNotFoundError(err, httpResp)
	}

	return c.outputByokKeyDescription(cmd, key)
}
