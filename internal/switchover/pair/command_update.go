package pair

import (
	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a switchover pair.",
		Long:  "Update a switchover pair's display name. This is the only mutable field.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Rename switchover pair "sw-123456".`,
				Code: `confluent switchover pair update sw-123456 --display-name "prod-kafka-dr-renamed"`,
			},
		),
	}

	cmd.Flags().String("display-name", "", "A human-readable name for the switchover pair.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("display-name"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	displayName, err := cmd.Flags().GetString("display-name")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pair := switchoverv1.SwitchoverV1SwitchoverPairUpdateRequest{
		Spec: switchoverv1.SwitchoverV1SwitchoverPairUpdateRequestSpec{
			DisplayName: switchoverv1.PtrString(displayName),
		},
	}

	result, err := c.V2Client.UpdateSwitchoverPair(args[0], environmentId, pair)
	if err != nil {
		return err
	}

	return printSwitchoverPair(cmd, result)
}
