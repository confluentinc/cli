package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *peeringCommand) newPeeringListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display peering connections in the current environment.",
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
	return list.Print()
}

func getCloud(peering networkingv1.NetworkingV1Peering) (string, error) {
	cloud := peering.Spec.GetCloud()

	if cloud.NetworkingV1AwsPeering != nil {
		return "AWS", nil
	} else if cloud.NetworkingV1GcpPeering != nil {
		return "GCP", nil
	} else if cloud.NetworkingV1AzurePeering != nil {
		return "Azure", nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
}
