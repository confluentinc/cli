package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display networks in the current environment.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	networks, err := c.V2Client.ListNetworks(environmentId)
	if err != nil {
		return err
	}

	return c.printList(cmd, networks)
}

func (c *command) printList(cmd *cobra.Command, networks []networkingv1.NetworkingV1Network) error {
	list := output.NewList(cmd)
	for _, network := range networks {
		zones := network.Spec.GetZones()
		activeConnectionTypes := network.Status.GetActiveConnectionTypes().Items

		if output.GetFormat(cmd) == output.Human {
			list.Add(&humanOut{
				Id:                    network.GetId(),
				EnvironmentId:         network.Spec.Environment.GetId(),
				Name:                  network.Spec.GetDisplayName(),
				Cloud:                 network.Spec.GetCloud(),
				Region:                network.Spec.GetRegion(),
				Cidr:                  network.Spec.GetCidr(),
				Zones:                 strings.Join(zones, ", "),
				DnsResolution:         network.Spec.DnsConfig.GetResolution(),
				Phase:                 network.Status.GetPhase(),
				ActiveConnectionTypes: strings.Join(activeConnectionTypes, ", "),
			})
		} else {
			list.Add(&serializedOut{
				Id:                    network.GetId(),
				EnvironmentId:         network.Spec.Environment.GetId(),
				Name:                  network.Spec.GetDisplayName(),
				Cloud:                 network.Spec.GetCloud(),
				Region:                network.Spec.GetRegion(),
				Cidr:                  network.Spec.GetCidr(),
				Zones:                 zones,
				DnsResolution:         network.Spec.DnsConfig.GetResolution(),
				Phase:                 network.Status.GetPhase(),
				ActiveConnectionTypes: activeConnectionTypes,
			})
		}
	}

	list.Filter([]string{"Id", "EnvironmentId", "Name", "Cloud", "Region", "Cidr", "Zones", "DnsResolution", "Phase", "ActiveConnectionTypes"})
	return list.Print()
}
