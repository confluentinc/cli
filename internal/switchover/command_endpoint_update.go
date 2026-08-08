package switchover

import (
	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a switchover endpoint.",
		Long:              "Update the display name of a switchover endpoint. Use `confluent switchover endpoint activate` to change the active endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validEndpointArgs),
		RunE:              c.endpointUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Rename switchover endpoint "se-123456".`,
				Code: `confluent switchover endpoint update se-123456 --name new-name`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the switchover endpoint.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) endpointUpdate(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	update := switchoverv1.SwitchoverV1SwitchoverEndpoint{
		Spec: &switchoverv1.SwitchoverV1SwitchoverEndpointSpec{
			DisplayName: switchoverv1.PtrString(name),
			Environment: &switchoverv1.EnvScopedObjectReference{Id: environmentId},
		},
	}

	endpoint, err := c.V2Client.UpdateSwitchoverEndpoint(args[0], update)
	if err != nil {
		return err
	}

	return printEndpointTable(cmd, endpoint)
}
