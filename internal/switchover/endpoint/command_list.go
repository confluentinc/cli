package endpoint

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List switchover endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List switchover endpoints in the current environment.",
				Code: "confluent switchover endpoint list",
			},
			examples.Example{
				Text: `List switchover endpoints for switchover pair "sw-123456".`,
				Code: "confluent switchover endpoint list --switchover-pair sw-123456",
			},
		),
	}

	cmd.Flags().String("switchover-pair", "", "Filter the results by the associated switchover pair ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	switchoverPair, err := cmd.Flags().GetString("switchover-pair")
	if err != nil {
		return err
	}

	endpoints, err := c.V2Client.ListSwitchoverEndpoints(environmentId, switchoverPair)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, endpoint := range endpoints {
		list.Add(&out{
			Id:             endpoint.GetId(),
			DisplayName:    endpoint.Spec.GetDisplayName(),
			SwitchoverPair: endpoint.Spec.GetSwitchoverPairId(),
			Environment:    endpoint.Spec.GetEnvironment(),
			Phase:          endpoint.Status.GetPhase(),
		})
	}
	return list.Print()
}
