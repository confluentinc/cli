package endpoint

import (
	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a switchover endpoint.",
		Long:  "Update a switchover endpoint's display name. This is the only mutable field.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Rename switchover endpoint "se-123456".`,
				Code: `confluent switchover endpoint update se-123456 --display-name "prod-kafka-dr-endpoint-renamed"`,
			},
		),
	}

	cmd.Flags().String("display-name", "", "A human-readable name for the switchover endpoint.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("display-name"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	displayName, err := cmd.Flags().GetString("display-name")
	if err != nil {
		return err
	}

	endpoint := switchoverv1.SwitchoverV1SwitchoverEndpointUpdateRequest{
		Spec: switchoverv1.SwitchoverV1SwitchoverEndpointUpdateRequestSpec{
			DisplayName: switchoverv1.PtrString(displayName),
		},
	}

	result, err := c.V2Client.UpdateSwitchoverEndpoint(args[0], endpoint)
	if err != nil {
		return err
	}

	return printSwitchoverEndpoint(cmd, result)
}
