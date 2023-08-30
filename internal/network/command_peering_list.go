package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *peeringCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List peering connections.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *peeringCommand) list(cmd *cobra.Command, _ []string) error {
	peerings, err := c.getPeerings()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, peering := range peerings {
		if peering.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if peering.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		cloud, err := getCloud(peering)
		if err != nil {
			return err
		}

		if output.GetFormat(cmd) == output.Human {
			list.Add(&peeringHumanOut{
				Id:        peering.GetId(),
				Name:      peering.Spec.GetDisplayName(),
				NetworkId: peering.Spec.Network.GetId(),
				Cloud:     cloud,
				Phase:     peering.Status.GetPhase(),
			})
		} else {
			list.Add(&peeringSerializedOut{
				Id:        peering.GetId(),
				Name:      peering.Spec.GetDisplayName(),
				NetworkId: peering.Spec.Network.GetId(),
				Cloud:     cloud,
				Phase:     peering.Status.GetPhase(),
			})
		}
	}
	list.Filter([]string{"Id", "Name", "NetworkId", "Cloud", "Phase"})
	return list.Print()
}
