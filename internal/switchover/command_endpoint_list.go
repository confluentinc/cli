package switchover

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List switchover endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.endpointList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	endpoints, err := c.V2Client.ListSwitchoverEndpoints(environmentId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, endpoint := range endpoints {
		if endpoint.Spec == nil {
			continue
		}
		out := &endpointOut{
			Id:             endpoint.GetId(),
			Name:           endpoint.Spec.GetDisplayName(),
			Environment:    endpoint.Spec.Environment.GetId(),
			SwitchoverPair: endpoint.Spec.SwitchoverPair.GetId(),
			Endpoints:      endpointNames(endpoint),
			Target:         endpoint.Spec.GetTarget(),
			DrEndpoint:     endpoint.Spec.GetDrEndpoint(),
		}
		if endpoint.Status != nil {
			out.Phase = endpoint.Status.GetPhase()
		}
		list.Add(out)
	}
	return list.Print()
}
