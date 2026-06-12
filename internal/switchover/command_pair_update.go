package switchover

import (
	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newPairUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a switchover pair.",
		Long:              "Update the display name of a switchover pair. Use `confluent switchover pair failover` to change the active member.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPairArgs),
		RunE:              c.pairUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Rename switchover pair "sw-123456".`,
				Code: `confluent switchover pair update sw-123456 --name new-name`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the switchover pair.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) pairUpdate(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	update := switchoverv1.SwitchoverV1SwitchoverPair{
		Spec: &switchoverv1.SwitchoverV1SwitchoverPairSpec{
			DisplayName: switchoverv1.PtrString(name),
			Environment: &switchoverv1.EnvScopedObjectReference{Id: environmentId},
		},
	}

	pair, err := c.V2Client.UpdateSwitchoverPair(args[0], update)
	if err != nil {
		return err
	}

	return printPairTable(cmd, pair)
}
